# expense-tracker
A simple CLI app for tracking expenses \
idea:https://roadmap.sh/projects/expense-tracker
### Usage:
```
expense-tracker [<operation>] [<parameters>]

operations:
    add
    --description <short description>
    --amount      <amount of expense>

    update
    --id          <existing id>
    --description <new short description>
    --amount      <new amount of expense>

    delete
    --id          <existing id>

    list

    summary
    --month       <month number>

    export

```