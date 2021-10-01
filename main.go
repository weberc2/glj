package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hokaccha/go-prettyjson"
	"github.com/itchyny/gojq"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

var (
	version string
)

func main() {
	app := &cli.App{
		Name:    "glj",
		Usage:   "JSON output for git log",
		Version: version,
		Authors: []*cli.Author{{
			Name:  "Craig Weber",
			Email: "weberc2@gmail.com",
		}},
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "query",
				Usage:   "A jq query",
				Aliases: []string{"q"},
			},
			&cli.StringFlag{
				Name:     "repo",
				Usage:    "The URL for the target git repository",
				Aliases:  []string{"r"},
				Required: true,
			},
			&cli.StringFlag{
				Name:      "key-file",
				Aliases:   []string{"k"},
				Usage:     "The private key file",
				Value:     "~/.ssh/id_rsa",
				Required:  false,
				TakesFile: true,
			},
		},
		Action: func(c *cli.Context) error {
			return run(
				os.Stdout,
				c.String("repo"),
				c.String("query"),
				c.String("key-file"),
			)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(output io.Writer, repo, query, privateKeyFile string) error {
	if strings.HasPrefix(privateKeyFile, "~/") {
		homeDir, err := homedir.Dir()
		if err != nil {
			return fmt.Errorf("Error loading home directory: %w", err)
		}
		privateKeyFile = strings.Replace(privateKeyFile, "~", homeDir, 1)
	}

	// Load the public keys
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		return fmt.Errorf(
			"Retrieving public keys from private key file: %w",
			err,
		)
	}

	// Clone the repo into memory
	r, err := git.Clone(
		memory.NewStorage(),
		nil,
		&git.CloneOptions{
			URL:      repo,
			Auth:     publicKeys,
			Progress: ioutil.Discard,
		},
	)
	if err != nil {
		return fmt.Errorf("Cloning repo: %w", err)
	}

	// Fetch the commit log
	commits, err := r.Log(&git.LogOptions{})
	if err != nil {
		return fmt.Errorf("Fetching log: %w", err)
	}

	// If there is no query, then skip all of the overhead associated with
	// querying
	if query == "" {
		for {
			c, err := commits.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("Fetching commit: %w", err)
			}

			if err := format(output, newCommit(c)); err != nil {
				return fmt.Errorf("Writing formatted commit: %w", err)
			}
		}
	} else {
		q, err := gojq.Parse(query)
		if err != nil {
			return fmt.Errorf("Parsing query: %w", err)
		}
		code, err := gojq.Compile(q)
		if err != nil {
			return fmt.Errorf("Compiling query: %w", err)
		}

		for {
			c, err := commits.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("Fetching commit: %w", err)
			}

			// GoJQ doesn't accept custom objects, but rather it expects us to
			// marshal an object to JSON, unmarshal it back into an
			// `interface{}`, and then pass *that* into the `Run()` function.
			data, err := json.Marshal(newCommit(c))
			if err != nil {
				return fmt.Errorf("Marshaling commit: %w", err)
			}

			var v interface{}
			if err := json.Unmarshal(data, &v); err != nil {
				return fmt.Errorf("Unmarshaling commit: %w", err)
			}

			it := code.Run(v)
			for v, ok := it.Next(); ok; v, ok = it.Next() {
				if err, ok := v.(error); ok {
					return fmt.Errorf("GOJQ error: %w", err)
				}

				if err := format(output, v); err != nil {
					return fmt.Errorf("Writing formatted commit: %w", err)
				}
			}
		}
	}
	return nil
}

// Same colors with gojq.
var formatter = &prettyjson.Formatter{
	KeyColor:    color.New(color.FgBlue, color.Bold),
	StringColor: color.New(color.FgGreen),
	BoolColor:   color.New(color.FgYellow),
	NumberColor: color.New(color.FgCyan),
	NullColor:   color.New(color.FgHiBlack),
	Indent:      2,
	Newline:     "\n",
}

// commit is just a friendlier structure for JSON marshaling. Specifically,
// hashes are represented as strings rather than arrays of integers.
type commit struct {
	Hash         string
	Author       object.Signature
	Committer    object.Signature
	PGPSignature string
	Message      string
	TreeHash     string
	ParentHashes []string
}

// create a `commit` from an `*object.Commit`
func newCommit(c *object.Commit) commit {
	parentHashes := make([]string, len(c.ParentHashes))
	for i := range c.ParentHashes {
		parentHashes[i] = c.ParentHashes[i].String()
	}
	return commit{
		Hash:         c.Hash.String(),
		Author:       c.Author,
		Committer:    c.Committer,
		PGPSignature: c.PGPSignature,
		Message:      c.Message,
		TreeHash:     c.TreeHash.String(),
		ParentHashes: parentHashes,
	}
}

// Use the `formatter` to marshal `v` and then write it to `output`.
func format(output io.Writer, v interface{}) error {
	data, err := formatter.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "%s\n", data)
	return err
}
