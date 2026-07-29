package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const OPERATIONS = "add update delete list summary export"
const DATA_FILE_PATH = "data.csv"
const EXPORT_FILE_PATH = "expenses.csv"

const MONTHS = " January February March April " +
	"May June Jule August " +
	"September October November December"

type Operation string

type Expense [4]string
type Expenses []Expense

func (exp Expense) GetSlice() []string {
	return exp[:]
}

func atoi(a string) int {
	ai, _ := strconv.Atoi(a)
	return ai
}

func itoa(a int) string {
	return strconv.Itoa(a)
}

func load(name string) (Expenses, error) {
	var (
		err    error
		exp    Expenses
		file   *os.File
		reader *csv.Reader
	)

	if _, err = os.Stat(name); os.IsNotExist(err) {
		return Expenses{
			{"ID", "Data", "Description", "Amount"},
			{"", "3", "5", "12"},
		}, nil
	}

	if file, err = os.Open(name); err != nil {
		return nil, err
	}
	defer file.Close()

	reader = csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		exp = append(exp, Expense(record))
	}
	return exp, nil
}

func save(name string, exp Expenses) error {
	var (
		file   *os.File
		writer *csv.Writer
		err    error
	)

	if file, err = os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666); err != nil {
		return err
	}
	defer file.Close()

	writer = csv.NewWriter(file)

	for i := 0; i < len(exp); i++ {
		err = writer.Write(exp[i].GetSlice())
		if err != nil {
			return err
		}
	}
	writer.Flush()

	return nil
}

func determineOperation(op string) (Operation, error) {
	for _, operation := range strings.Split(OPERATIONS, " ") {
		if op == operation {
			return Operation(op), nil
		}
	}
	return "", fmt.Errorf("Operation not found")
}

func exitWithError(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func (exp *Expenses) Print() {

	extData := (*exp)[len(*exp)-1]

	for _, e := range (*exp)[:len(*exp)-1] {
		fmt.Println(
			e[0] + strings.Repeat(" ", atoi(extData[1])-len(e[0])) +
				e[1] + strings.Repeat(" ", atoi(extData[2])-len(e[1])) +
				e[2] + strings.Repeat(" ", atoi(extData[3])-len([]rune(e[2]))) +
				e[3],
		)
	}
}

func (exp *Expenses) newLastIndex() string {
	e := *exp
	elen := len(e)

	if elen == 2 {
		return "1"
	}

	indexInt, _ := strconv.Atoi(e[elen-2][0])
	indexInt++
	return strconv.Itoa(indexInt)
}

func (exp *Expenses) findSliceIndex(id int) (int, error) {
	e := *exp
	idString := strconv.Itoa(id)
	for i := 0; i < len(e); i++ {
		if e[i][0] == idString {
			return i, nil
		}
	}
	return 0, fmt.Errorf("error: id %d is not exist or out of range", id)
}

func (exp *Expenses) updateIndents(curExp Expense) {

	lastIndex := len(*exp) - 1
	extData := (*exp)[lastIndex]

	(*exp)[lastIndex] = Expense{
		"",
		itoa(max(atoi(extData[1]), len(curExp[0])+1)),         // indent after ID
		itoa(max(atoi(extData[2]), len(curExp[1])+1)),         // indent after Date
		itoa(max(atoi(extData[3]), len([]rune(curExp[2]))+1)), // indent after Description
	}
}

func (exp *Expenses) add(args []string) error {
	if len(args) < 5 {
		return fmt.Errorf("usage: expense-tracker add --description <desc> --amount <amount>")
	}

	addFlags := flag.NewFlagSet("add", flag.ExitOnError)
	description := addFlags.String("description", "", "description of your expense")
	amount := addFlags.Int("amount", 0, "amount of your expense")

	addFlags.Parse(args[1:])

	if *description == "" || *amount == 0 {
		return fmt.Errorf("usage: expense-tracker add --description <desc> --amount <amount>\n" +
			"                                   string^^^^^^       int^^^^^^^^")
	}

	*exp = append(*exp, Expense{
		exp.newLastIndex(),
		time.Now().Format(time.DateOnly),
		*description,
		strconv.Itoa(*amount),
	})

	lastIndex := len(*exp) - 1

	(*exp)[lastIndex], (*exp)[lastIndex-1] =
		(*exp)[lastIndex-1], (*exp)[lastIndex]

	exp.updateIndents((*exp)[lastIndex-1])

	if err := save(DATA_FILE_PATH, *exp); err != nil {
		return err
	}

	fmt.Printf("Expense added successfully (ID: %d)\n", lastIndex-1)
	return nil
}

func (exp *Expenses) update(args []string) error {
	var (
		trueID int
		err    error
	)

	if len(args) < 7 {
		return fmt.Errorf("usage: expense-tracker update " +
			"--id <existing id> --description <desc> --amount <amount>")
	}

	updateFlags := flag.NewFlagSet("update", flag.ExitOnError)
	id := updateFlags.Int("id", 0, "id of existing expense")
	description := updateFlags.String("description", "", "description of your expense")
	amount := updateFlags.Int("amount", 0, "amount of your expense")

	updateFlags.Parse(args[1:])

	if *description == "" || *amount <= 0 || *id <= 0 {
		return fmt.Errorf("usage: expense-tracker update " +
			"--id <existing id> --description <desc> --amount <amount>\n" +
			"                                " +
			"int^^^^^^^^^^^^^         string^^^^^^       int^^^^^^^^")
	}

	if trueID, err = exp.findSliceIndex(*id); err != nil {
		return err
	}

	(*exp)[trueID] = Expense{
		strconv.Itoa(trueID),
		time.Now().Format(time.DateOnly),
		*description,
		strconv.Itoa(*amount),
	}

	exp.updateIndents((*exp)[trueID])

	if err := save(DATA_FILE_PATH, *exp); err != nil {
		return err
	}

	fmt.Println("Expense updated successfully")
	return nil
}

func (exp *Expenses) delete(args []string) error {
	var (
		trueID int
		err    error
	)

	if len(args) < 3 {
		return fmt.Errorf("usage: expense-tracker update " +
			"--id <existing id>")
	}
	deleteFlags := flag.NewFlagSet("update", flag.ExitOnError)
	id := deleteFlags.Int("id", 0, "id of existing expense")

	deleteFlags.Parse(args[1:])

	if *id <= 0 {
		return fmt.Errorf("usage: expense-tracker update" +
			" --id <existing id>\n" +
			"   int^^^^^^^^^^^^^")
	}

	if trueID, err = exp.findSliceIndex(*id); err != nil {
		return err
	}

	*exp = append((*exp)[:trueID], (*exp)[trueID+1:]...)

	if err := save(DATA_FILE_PATH, *exp); err != nil {
		return err
	}

	fmt.Println("Expense deleted successfully")
	return nil
}

func (exp *Expenses) list() error {
	(*exp).Print()
	return nil
}

func (exp *Expenses) summary(args []string) error {
	sumFlags := flag.NewFlagSet("summary", flag.ExitOnError)

	month := sumFlags.Int("month", 0, "filter expenses by categories")

	sumFlags.Parse(args[1:])

	if m := *month; m < 0 || m > 12 {
		return fmt.Errorf("error: month is out of existing months [%d]", m)
	}

	sum := 0
	if *month == 0 {
		for i := 1; i < len(*exp)-1; i++ {
			e := (*exp)[i]
			sum += atoi(e[3])
		}
		fmt.Printf("Total expenses: $%d\n", sum)
	} else {
		for i := 1; i < len(*exp)-1; i++ {
			e := (*exp)[i]
			if *month == atoi(e[1][6:7]) {
				sum += atoi(e[3])
			}
		}
		fmt.Printf("Total expenses for %s: $%d\n",
			strings.Split(MONTHS, " ")[*month], sum)
	}
	return nil
}

func (op Operation) Process(exp *Expenses, args []string) (err error) {
	switch op {
	case "add":
		err = exp.add(args)
	case "update":
		err = exp.update(args)
	case "delete":
		err = exp.delete(args)
	case "list":
		err = exp.list()
	case "summary":
		err = exp.summary(args)
	case "export":
		err = save(EXPORT_FILE_PATH, (*exp)[:len(*exp)-1])
	}
	return err
}

func main() {
	var (
		op  Operation
		exp Expenses
		err error
	)

	if len(os.Args) < 2 {
		exitWithError(fmt.Errorf("error: app must have an operation"))
	}

	args := os.Args[1:]

	if op, err = determineOperation(args[0]); err != nil {
		exitWithError(err)
	}

	if exp, err = load(DATA_FILE_PATH); err != nil {
		exitWithError(err)
	}

	if err = op.Process(&exp, args); err != nil {
		exitWithError(err)
	}
}
