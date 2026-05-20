package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const anvilArt = `

 ________  ________  ________ ________  ________  ________  _______      
|\   __  \|\   ___ \|\  _____\\   __  \|\   __  \|\   ____\|\  ___ \     
\ \  \|\  \ \  \_|\ \ \  \__/\ \  \|\  \ \  \|\  \ \  \___|\ \   __/|    
 \ \   ____\ \  \ \\ \ \   __\\ \  \\\  \ \   _  _\ \  \  __\ \  \_|/__  
  \ \  \___|\ \  \_\\ \ \  \_| \ \  \\\  \ \  \\  \\ \  \|\  \ \  \_|\ \ 
   \ \__\    \ \_______\ \__\   \ \_______\ \__\\ _\\ \_______\ \_______\
    \|__|     \|_______|\|__|    \|_______|\|__|\|__|\|_______|\|_______|  
                                                                                                                                                                                                                
		⠀⠀⠀⠀⠀⠀⠀⢰⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⡄⠀⠀⠀⠀⠀
		⠀⠹⣿⣿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢠⣄⡀⠀⠀
		⠀⠀⠙⢿⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⢸⣿⣿⡶⠀
		⠀⠀⠀⠀⠉⠛⠇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠸⠟⠋⠀⠀
		⠀⠀⠀⠀⠀⠀⠀⠸⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠿⠇⠀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣶⣶⣶⣶⣶⣶⣶⣶⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣷⡀⠀⠀⠀⠀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣄⠀⠀⠀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⣀⣀⣈⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣉⣁⣀⣀⠀⠀⠀⠀
		⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⠀⠀⠀
`

const rootHelpTemplate = `{{pdArt}}

{{pdTitle "PDFORGE"}}
{{pdMuted "Local, privacy-first PDF toolkit"}}

{{pdTitle "USAGE"}}
	{{.UseLine}}

{{if .Long}}{{pdTitle "DESCRIPTION"}}
{{.Long}}
{{end}}{{if .HasAvailableSubCommands}}{{pdTitle "COMMANDS"}}
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}{{if not .Hidden}}  {{pdAccent (rpad .Name .NamePadding)}} {{.Short}}
{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
{{pdTitle "FLAGS"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
{{pdTitle "GLOBAL FLAGS"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .Example}}

{{pdTitle "EXAMPLES"}}
{{.Example}}{{end}}

{{pdMuted "Run pdforge [command] --help for detailed command usage."}}
`

const subHelpTemplate = `{{pdTitle "USAGE"}}
	{{.UseLine}}

{{if .Long}}{{pdTitle "DESCRIPTION"}}
{{.Long}}
{{end}}{{if .HasAvailableSubCommands}}{{pdTitle "COMMANDS"}}
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}{{if not .Hidden}}  {{pdAccent (rpad .Name .NamePadding)}} {{.Short}}
{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
{{pdTitle "FLAGS"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}
{{pdTitle "GLOBAL FLAGS"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .Example}}

{{pdTitle "EXAMPLES"}}
{{.Example}}{{end}}

{{pdMuted "Run pdforge [command] --help for detailed command usage."}}
`


var enableANSI = shouldUseColor()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pdforge",
	Short: "pdforge: local, privacy-first PDF toolkit",
	Long: `pdforge is a local, open-source CLI for common PDF workflows.
All processing happens on your machine, with no uploads and no cloud dependency.
`,
	Example: `	pdforge convert image.jpg
	pdforge merge a.pdf b.pdf
	pdforge split report.pdf --page 1-3
	pdforge rmpage report.pdf 8
	pdforge optimize large.pdf`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func shouldUseColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fi.Mode() & os.ModeCharDevice) != 0
}

func paint(s, code string) string {
	if !enableANSI {
		return s
	}

	return code + s + "\x1b[0m"
}

func init() {
	cobra.AddTemplateFunc("pdArt", func() string {
		return anvilArt
	})

	cobra.AddTemplateFunc("pdTitle", func(s string) string {
		return paint(s, "\x1b[1;36m")
	})

	cobra.AddTemplateFunc("pdAccent", func(s string) string {
		return paint(s, "\x1b[1;37m")
	})

	cobra.AddTemplateFunc("pdMuted", func(s string) string {
		return paint(s, "\x1b[90m")
	})

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceUsage = true
	rootCmd.SetHelpTemplate(rootHelpTemplate)
}
