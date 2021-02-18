package add

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cli/cli/internal/ghinstance"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
)

type AddOptions struct {
	IO         *iostreams.IOStreams
	HTTPClient func() (*http.Client, error)

	KeyFile string
	Title   string
}

func NewCmdAdd(f *cmdutil.Factory, runF func(*AddOptions) error) *cobra.Command {
	opts := &AddOptions{
		HTTPClient: f.HttpClient,
		IO:         f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "add [<key-file>]",
		Short: "Add an SSH key to your GitHub account",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				if opts.IO.IsStdoutTTY() && opts.IO.IsStdinTTY() {
					return &cmdutil.FlagError{Err: errors.New("public key file missing")}
				}
				opts.KeyFile = "-"
			} else {
				opts.KeyFile = args[0]
			}

			if runF != nil {
				return runF(opts)
			}
			return runAdd(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Title, "title", "t", "", "Title for the new key")

	cmdutil.DeferCompletion(func() {
		carapace.Gen(cmd).PositionalCompletion(
			carapace.ActionFiles(),
		)
	})

	return cmd
}

func runAdd(opts *AddOptions) error {
	httpClient, err := opts.HTTPClient()
	if err != nil {
		return err
	}

	var keyReader io.Reader
	if opts.KeyFile == "-" {
		keyReader = opts.IO.In
		defer opts.IO.In.Close()
	} else {
		f, err := os.Open(opts.KeyFile)
		if err != nil {
			return err
		}
		defer f.Close()
		keyReader = f
	}

	hostname := ghinstance.OverridableDefault()
	err = SSHKeyUpload(httpClient, hostname, keyReader, opts.Title)
	if err != nil {
		if errors.Is(err, scopesError) {
			cs := opts.IO.ColorScheme()
			fmt.Fprint(opts.IO.ErrOut, "Error: insufficient OAuth scopes to list SSH keys\n")
			fmt.Fprintf(opts.IO.ErrOut, "Run the following to grant scopes: %s\n", cs.Bold("gh auth refresh -s write:public_key"))
			return cmdutil.SilentError
		}
		return err
	}

	if opts.IO.IsStdoutTTY() {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut, "%s Public key added to your account\n", cs.SuccessIcon())
	}
	return nil
}
