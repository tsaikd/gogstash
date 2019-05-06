gogstash mutate filter module
=============================

## Synopsis

At least one of split, replace or merge need to be set
```yaml
filter:
  - type: mutate
    # 2 item array. First item is field to split. Second item is string to split with
    split: ["mycommalistvalue", ","]
    # 3 item array. First item is field to modify. Second item is value to replace. Third item is new value to replace with.
    replace: ["myfield", "oldvalue", "newvalue"]
    # 2 item array. First item is field to merge on. Second item is string to add to merged field
    merge: ["listofthings", "newitemtoadd"]
```
