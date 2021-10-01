# Git Log JSON

`glj` is a CLI for analyzing the git log as JSON, including `jq`-esque query
support.

## USAGE

```bash
# Select commits authored after 2021-10-01 and simplify the output
$ glj \
    --repo git@github.com:weberc2/glj.git \
    --query 'select(.Author.When > "2021-10-01") |
        {Author: .Author.Name, Date: .Author.When, Message: .Message}'
{
  "Author": "Craig Weber",
  "Date": "2021-10-01T15:27:39-05:00",
  "Message": "Added README\n"
}
{
  "Author": "Craig Weber",
  "Date": "2021-10-01T15:13:39-05:00",
  "Message": "Fix cli version\n"
}
{
  "Author": "Craig Weber",
  "Date": "2021-10-01T14:34:38-05:00",
  "Message": "Integrate gojq query support\n"
}
```
