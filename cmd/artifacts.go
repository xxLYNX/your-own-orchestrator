package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"text/tabwriter"

	"yoo/internal/config"
	"yoo/internal/database"
	"yoo/internal/models"

	"github.com/spf13/cobra"
)

var (
	artifactType        string
	artifactName        string
	artifactFile        string
	artifactURL         string
	artifactDescription string
	artifactRequired    bool
)

// artifactsCmd represents the artifacts command
var artifactsCmd = &cobra.Command{
	Use:     "artifacts",
	Aliases: []string{"artifact", "art"},
	Short:   "Manage artifacts (inputs and outputs) for templated notes",
	Long: `Manage artifacts for templated notes.

Artifacts are inputs and outputs associated with a templated note. They can be:
- Files: Local file paths that will be stored as absolute paths
- URLs: Web links to resources
- Folders: Directory paths
- Text: Text content

Examples:
  yoo artifact list 42
  yoo artifact add 42 --type input --name source --file ./data.csv
  yoo artifact add 42 --type output --name report --url https://example.com/report
  yoo artifact show 42 source
  yoo artifact delete 42 source
  yoo artifact open 42 report`,
}

// artifactsListCmd lists all artifacts for a note
var artifactsListCmd = &cobra.Command{
	Use:   "list <note-id>",
	Short: "List all artifacts for a templated note",
	Long: `List all artifacts (inputs and outputs) for a templated note.

Artifacts are shown grouped by type (Inputs and Outputs), with indicators
for required artifacts and whether they have been provided.

Example:
  yoo artifact list 42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note template
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated or not found: %w", err)
		}

		// List artifacts
		artifacts, err := database.ListArtifacts(db.Conn(), noteTemplate.ID)
		if err != nil {
			return fmt.Errorf("failed to list artifacts: %w", err)
		}

		if len(artifacts) == 0 {
			fmt.Println("No artifacts defined for this note.")
			return nil
		}

		// Group by artifact type
		inputs := []*models.Artifact{}
		outputs := []*models.Artifact{}
		for _, art := range artifacts {
			if art.ArtifactType == "input" {
				inputs = append(inputs, art)
			} else {
				outputs = append(outputs, art)
			}
		}

		// Display inputs
		if len(inputs) > 0 {
			fmt.Println("Inputs:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, art := range inputs {
				status := " "
				if art.Value != "" {
					status = "✓"
				}
				required := ""
				if art.Required {
					required = " [required]"
				}
				fmt.Fprintf(w, "  %s\t%s\t%s\t%s%s\n", status, art.Name, art.Type, art.Description, required)
			}
			w.Flush()
			fmt.Println()
		}

		// Display outputs
		if len(outputs) > 0 {
			fmt.Println("Outputs:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, art := range outputs {
				status := " "
				if art.Value != "" {
					status = "✓"
				}
				required := ""
				if art.Required {
					required = " [required]"
				}
				fmt.Fprintf(w, "  %s\t%s\t%s\t%s%s\n", status, art.Name, art.Type, art.Description, required)
			}
			w.Flush()
			fmt.Println()
		}

		return nil
	},
}

// artifactsAddCmd adds a new artifact
var artifactsAddCmd = &cobra.Command{
	Use:   "add <note-id>",
	Short: "Add an artifact to a templated note",
	Long: `Add an artifact (input or output) to a templated note.

You must specify either --file or --url to define the artifact value.
For files, the path will be validated and stored as an absolute path.

Examples:
  yoo artifact add 42 --type input --name source --file ./data.csv --description "Source data"
  yoo artifact add 42 --type output --name report --url https://example.com/report
  yoo artifact add 42 --type input --name config --file ./config.json --required`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		// Validate required flags
		if artifactType == "" {
			return fmt.Errorf("--type flag is required (input or output)")
		}
		if artifactType != "input" && artifactType != "output" {
			return fmt.Errorf("artifact type must be 'input' or 'output', got: %s", artifactType)
		}
		if artifactName == "" {
			return fmt.Errorf("--name flag is required")
		}
		if artifactFile == "" && artifactURL == "" {
			return fmt.Errorf("either --file or --url must be specified")
		}
		if artifactFile != "" && artifactURL != "" {
			return fmt.Errorf("cannot specify both --file and --url")
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note template
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated or not found: %w", err)
		}

		// Determine artifact type and value
		var artType, value string
		if artifactFile != "" {
			// Validate file exists
			if _, err := os.Stat(artifactFile); os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", artifactFile)
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(artifactFile)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}

			artType = "file"
			value = absPath
		} else {
			artType = "url"
			value = artifactURL
		}

		// Create artifact
		artifact := &models.Artifact{
			NoteTemplateID: noteTemplate.ID,
			ArtifactType:   artifactType,
			Name:           artifactName,
			Type:           artType,
			Value:          value,
			Description:    artifactDescription,
			Required:       artifactRequired,
		}

		if err := database.CreateArtifact(db.Conn(), artifact); err != nil {
			return fmt.Errorf("failed to create artifact: %w", err)
		}

		fmt.Printf("✓ Added %s artifact '%s' (%s)\n", artifactType, artifactName, artType)
		if artifactFile != "" {
			fmt.Printf("  Path: %s\n", value)
		} else {
			fmt.Printf("  URL: %s\n", value)
		}

		return nil
	},
}

// artifactsShowCmd shows details of a specific artifact
var artifactsShowCmd = &cobra.Command{
	Use:   "show <note-id> <name>",
	Short: "Show artifact details",
	Long: `Show detailed information about a specific artifact.

Example:
  yoo artifact show 42 source`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}
		name := args[1]

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note template
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated or not found: %w", err)
		}

		// Get artifact
		artifact, err := database.GetArtifact(db.Conn(), noteTemplate.ID, name)
		if err != nil {
			return err
		}

		// Display artifact details
		fmt.Printf("Name: %s\n", artifact.Name)
		fmt.Printf("Artifact Type: %s\n", artifact.ArtifactType)
		fmt.Printf("Type: %s\n", artifact.Type)
		fmt.Printf("Value: %s\n", artifact.Value)
		if artifact.Description != "" {
			fmt.Printf("Description: %s\n", artifact.Description)
		}
		fmt.Printf("Required: %v\n", artifact.Required)
		fmt.Printf("Created: %s\n", artifact.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", artifact.UpdatedAt.Format("2006-01-02 15:04:05"))

		// Check if file exists
		if artifact.Type == "file" {
			if _, err := os.Stat(artifact.Value); os.IsNotExist(err) {
				fmt.Printf("\n⚠ Warning: File does not exist at path: %s\n", artifact.Value)
			} else {
				fmt.Printf("\n✓ File exists\n")
			}
		}

		return nil
	},
}

// artifactsDeleteCmd deletes an artifact
var artifactsDeleteCmd = &cobra.Command{
	Use:   "delete <note-id> <name>",
	Short: "Delete an artifact",
	Long: `Delete an artifact from a templated note.

This only removes the artifact reference from the database,
it does not delete the actual file or URL.

Example:
  yoo artifact delete 42 source`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}
		name := args[1]

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note template
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated or not found: %w", err)
		}

		// Delete artifact
		if err := database.DeleteArtifact(db.Conn(), noteTemplate.ID, name); err != nil {
			return err
		}

		fmt.Printf("✓ Deleted artifact '%s'\n", name)
		return nil
	},
}

// artifactsOpenCmd opens an artifact (file or URL)
var artifactsOpenCmd = &cobra.Command{
	Use:   "open <note-id> <name>",
	Short: "Open an artifact (file or URL)",
	Long: `Open an artifact with the default system application.

For files, this opens the file with the default application.
For URLs, this opens the URL in the default web browser.

Example:
  yoo artifact open 42 report`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}
		name := args[1]

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note template
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated or not found: %w", err)
		}

		// Get artifact
		artifact, err := database.GetArtifact(db.Conn(), noteTemplate.ID, name)
		if err != nil {
			return err
		}

		// Check value exists
		if artifact.Value == "" {
			return fmt.Errorf("artifact '%s' has no value set", name)
		}

		// For files, check if it exists
		if artifact.Type == "file" {
			if _, err := os.Stat(artifact.Value); os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", artifact.Value)
			}
		}

		// Open with platform-specific command
		if err := openWithDefaultApp(artifact.Value); err != nil {
			return fmt.Errorf("failed to open artifact: %w", err)
		}

		fmt.Printf("✓ Opened %s: %s\n", artifact.Type, artifact.Value)
		return nil
	},
}

// openWithDefaultApp opens a file or URL with the default system application
func openWithDefaultApp(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default: // linux and other unix-like systems
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}

func init() {
	rootCmd.AddCommand(artifactsCmd)

	// Add subcommands
	artifactsCmd.AddCommand(artifactsListCmd)
	artifactsCmd.AddCommand(artifactsAddCmd)
	artifactsCmd.AddCommand(artifactsShowCmd)
	artifactsCmd.AddCommand(artifactsDeleteCmd)
	artifactsCmd.AddCommand(artifactsOpenCmd)

	// Flags for add command
	artifactsAddCmd.Flags().StringVar(&artifactType, "type", "", "Artifact type (input or output)")
	artifactsAddCmd.Flags().StringVar(&artifactName, "name", "", "Artifact name")
	artifactsAddCmd.Flags().StringVar(&artifactFile, "file", "", "File path")
	artifactsAddCmd.Flags().StringVar(&artifactURL, "url", "", "URL")
	artifactsAddCmd.Flags().StringVar(&artifactDescription, "description", "", "Artifact description")
	artifactsAddCmd.Flags().BoolVar(&artifactRequired, "required", false, "Mark artifact as required")
}
