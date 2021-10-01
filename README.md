# Git Log JSON

`glj` is a CLI for analyzing the git log as JSON, including `jq`-esque query
support.

## USAGE

```bash
$ glj --repo git@github.com:weberc2/glj.git --query 'select(.Author.When > "2021-10-01")'
{
  ![#0000FF]("Author"): {
    ![#0000FF]("Email"): ![#00FF00]("weberc2@gmail.com"),
    ![#0000FF]("Name"): ![#00FF00]("Craig Weber"),
    ![#0000FF]("When"): ![#00FF00]("2021-10-01T14:34:38-05:00"
  },
  ![#0000FF]("Committer"): {
    ![#0000FF]("Email"): ![#00FF00]("weberc2@gmail.com"),
    ![#0000FF]("Name"): ![#00FF00]("Craig Weber"),
    ![#0000FF]("When"): ![#00FF00]("2021-10-01T14:34:38-05:00"
  },
  ![#0000FF]("Hash"): ![#00FF00]("a93a01bb1020e2cec5dc2e01ea2573cad9656f48"),
  ![#0000FF]("Message"): ![#00FF00]("Integrate gojq query support\n"),
  ![#0000FF]("PGPSignature"): ![#00FF00](""),
  ![#0000FF]("ParentHashes"): [
    "![#00FF00](ab5c37bda80f5d4acb81fc013152cb6c1ecfdc55"
  ],
  ![#0000FF]("TreeHash"): ![#00FF00]("0cac94407820f42488ab3dd57e728eed39602d4f"
}
```
