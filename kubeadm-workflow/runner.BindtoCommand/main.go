package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	// Import actual cobra and pflag
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// --- Start Simulation of the 'workflow' package ---
// (Adapted and expanded from previous examples based on BindToCommand code)

// RunData interface remains the same
type RunData interface{}

// Phase definition (simplified for adding to Runner)
type Phase struct {
	Name         string
	Short        string
	Long         string   // Added for phase help
	Example      string   // Added for phase help
	Aliases      []string // Added for phase help
	Hidden       bool     // To test phase visibility
	Run          func(data RunData) error
	InheritFlags []string       // Flags to inherit from parent command
	LocalFlags   *pflag.FlagSet // Flags specific to this phase
	// --- Fields needed for BindToCommand's internal logic ---
	RunAllSiblings bool                 // For phase group execution
	ArgsValidator  cobra.PositionalArgs // Args validation for the phase
	SubPhases      []Phase              // For hierarchical phases
}

// Internal representation used by BindToCommand (simulated)
type phaseRunner struct {
	// Copied from Phase definition
	Name           string
	Short          string
	Long           string
	Example        string
	Aliases        []string
	Hidden         bool
	Run            func(data RunData) error
	InheritFlags   []string
	LocalFlags     *pflag.FlagSet
	RunAllSiblings bool
	ArgsValidator  cobra.PositionalArgs
	// Internal fields set during prepareForExecution/visitAll
	Phases        []*phaseRunner // Child phaseRunners
	parent        *phaseRunner   // Parent phaseRunner
	level         int            // Depth in the hierarchy
	generatedName string         // Unique identifier (e.g., "setup/download")
	runner        *Runner        // Back pointer to the main runner
}

// Options for the workflow execution
type WorkflowOptions struct {
	SkipPhases   []string
	FilterPhases []string // Used internally when running a specific phase subcommand
}

// Runner manages and executes a sequence of phases.
type Runner struct {
	// User-added phases (raw structure)
	userPhases []Phase
	// Internal representation built from userPhases
	Phases []*phaseRunner // Hierarchical internal representation
	// Other fields
	dataInitializer    func(cmd *cobra.Command, args []string) (RunData, error)
	runCmd             *cobra.Command // The command that triggered the Run (set by BindToCommand/RunE)
	Options            *WorkflowOptions
	cmdAdditionalFlags *pflag.FlagSet // Extra flags to potentially add to phases
}

// NewRunner creates a new workflow Runner.
func NewRunner() *Runner {
	return &Runner{
		userPhases:         make([]Phase, 0),
		Phases:             make([]*phaseRunner, 0),
		Options:            &WorkflowOptions{},
		cmdAdditionalFlags: pflag.NewFlagSet("additional", pflag.ContinueOnError), // Initialize example additional flags
	}
}

// AppendPhase adds a top-level phase definition.
func (e *Runner) AppendPhase(p Phase) {
	e.userPhases = append(e.userPhases, p)
}

// SetDataInitializer sets the function used to create the initial RunData.
func (e *Runner) SetDataInitializer(initializer func(cmd *cobra.Command, args []string) (RunData, error)) {
	e.dataInitializer = initializer
}

// --- Simulation of methods used by BindToCommand ---

// prepareForExecution: Builds the internal phaseRunner tree from userPhases.
// This is a simplified simulation assuming userPhases defines the structure.
func (e *Runner) prepareForExecution() {
	fmt.Println("[Runner.prepareForExecution called]") // Indicate it's running
	e.Phases = make([]*phaseRunner, 0)                 // Reset internal phases
	nameMap := make(map[string]*phaseRunner)           // To link parents

	var buildTree func([]Phase, *phaseRunner, int) []*phaseRunner
	buildTree = func(phaseDefs []Phase, parent *phaseRunner, level int) []*phaseRunner {
		runners := make([]*phaseRunner, 0)
		for _, pDef := range phaseDefs {
			pr := &phaseRunner{
				Name:           pDef.Name,
				Short:          pDef.Short,
				Long:           pDef.Long,
				Example:        pDef.Example,
				Aliases:        pDef.Aliases,
				Hidden:         pDef.Hidden,
				Run:            pDef.Run,
				InheritFlags:   pDef.InheritFlags,
				LocalFlags:     pDef.LocalFlags,
				RunAllSiblings: pDef.RunAllSiblings,
				ArgsValidator:  pDef.ArgsValidator,
				parent:         parent,
				level:          level,
				runner:         e,
			}
			// Generate a simple unique name (adjust for real nesting)
			if parent != nil {
				pr.generatedName = parent.generatedName + "/" + strings.ToLower(pDef.Name)
			} else {
				pr.generatedName = strings.ToLower(pDef.Name)
			}

			// Recursively build for sub-phases
			if len(pDef.SubPhases) > 0 {
				pr.Phases = buildTree(pDef.SubPhases, pr, level+1)
			}

			nameMap[pr.generatedName] = pr
			runners = append(runners, pr)
		}
		return runners
	}

	e.Phases = buildTree(e.userPhases, nil, 0)
}

// visitAll: Traverses the internal phaseRunner tree.
func (e *Runner) visitAll(visitor func(p *phaseRunner) error) error {
	var visitNode func(*phaseRunner) error
	visitNode = func(p *phaseRunner) error {
		if err := visitor(p); err != nil {
			return err // Stop traversal on error
		}
		for _, child := range p.Phases {
			if err := visitNode(child); err != nil {
				return err
			}
		}
		return nil
	}

	for _, rootPhase := range e.Phases {
		if err := visitNode(rootPhase); err != nil {
			return err
		}
	}
	return nil
}

// Help: Generates phase list help text (simplified).
func (e *Runner) Help(cmdUse string) string {
	var sb strings.Builder
	sb.WriteString("Available phases:\n")

	e.visitAll(func(p *phaseRunner) error { // Use visitAll to traverse
		if p.Hidden {
			return nil // Skip hidden phases
		}
		indent := strings.Repeat("  ", p.level*2)
		sb.WriteString(fmt.Sprintf("%s* %s: %s\n", indent, p.Name, p.Short))
		return nil
	})

	sb.WriteString(fmt.Sprintf("\nYou can run a specific phase using '%s phase <phase-name>'.", cmdUse))
	return sb.String()
}

// inheritsFlags: Simulates copying flags (simplified).
func inheritsFlags(sourceFlags *pflag.FlagSet, destFlags *pflag.FlagSet, inheritNames []string) {
	if sourceFlags == nil || destFlags == nil {
		return
	}
	// If inheritNames is nil/empty, inherit all. Otherwise, only inherit specified ones.
	inheritAll := len(inheritNames) == 0
	allowedInherit := make(map[string]bool)
	if !inheritAll {
		for _, name := range inheritNames {
			allowedInherit[name] = true
		}
	}

	sourceFlags.VisitAll(func(flag *pflag.Flag) {
		// Don't inherit the help flag or flags already defined in dest
		if flag.Name == "help" || destFlags.Lookup(flag.Name) != nil {
			return
		}
		if inheritAll || allowedInherit[flag.Name] {
			destFlags.AddFlag(flag) // Add a copy
		}
	})
}

// --- The actual BindToCommand function (from user input) ---
func (e *Runner) BindToCommand(cmd *cobra.Command) {
	// keep track of the command triggering the runner
	e.runCmd = cmd // Keep original trigger command initially

	e.prepareForExecution() // Make sure internal structure is ready

	// return early if no phases prepared (prepareForExecution should populate e.Phases)
	if len(e.Phases) == 0 {
		fmt.Println("[BindToCommand] No phases found after preparation.")
		return
	}

	// adds the phases subcommand
	phaseCommand := &cobra.Command{
		Use:   "phase",
		Short: fmt.Sprintf("Use this command to invoke single phases of the %s workflow", cmd.Name()),
		// Prevent cobra from adding implicit completion command for 'phase' itself
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	cmd.AddCommand(phaseCommand)

	// generate all the nested subcommands for invoking single phases
	subcommands := map[string]*cobra.Command{} // Map generatedName to cobra.Command

	e.visitAll(func(p *phaseRunner) error { // Iterate using the internal structure
		// skip hidden phases
		if p.Hidden {
			return nil
		}

		// initialize phase selector (unique name used for filtering)
		phaseSelector := p.generatedName

		// if requested, set the phase to run all the sibling phases
		// Note: This requires parent to be non-nil, root phases can't run siblings this way
		if p.RunAllSiblings && p.parent != nil {
			phaseSelector = p.parent.generatedName // Target the parent group
		}

		// creates phase subcommand
		phaseCmd := &cobra.Command{
			Use:     strings.ToLower(p.Name), // Use the simple name for the subcommand
			Short:   p.Short,
			Long:    p.Long,
			Example: p.Example,
			Aliases: p.Aliases,
			RunE: func(phaseSubCmd *cobra.Command, args []string) error {
				fmt.Printf("[RunE for phase '%s' triggered]\n", p.Name)
				// if the phase has subphases, just print the help and exit
				if len(p.Phases) > 0 {
					fmt.Printf("Phase '%s' has subphases. Use a subphase or run the main command.\n", p.Name)
					return phaseSubCmd.Help()
				}

				// overrides the command triggering the Runner using the phaseCmd
				e.runCmd = phaseSubCmd                           // THIS command triggered the run
				e.Options.FilterPhases = []string{phaseSelector} // Tell Run to execute only this phase/group
				fmt.Printf("  Filtering execution to: %v\n", e.Options.FilterPhases)
				// Reset skip phases when running a specific phase directly
				e.Options.SkipPhases = nil
				return e.Run(args) // Call the main runner's Run method
			},
		}

		// makes the new command inherits local flags from the *original* parent command (cmd)
		inheritsFlags(cmd.Flags(), phaseCmd.Flags(), p.InheritFlags)

		// makes the new command inherits additional flags for phases (if defined)
		if e.cmdAdditionalFlags != nil {
			inheritsFlags(e.cmdAdditionalFlags, phaseCmd.Flags(), p.InheritFlags)
		}

		// If defined, add phase local flags
		if p.LocalFlags != nil {
			p.LocalFlags.VisitAll(func(f *pflag.Flag) {
				phaseCmd.Flags().AddFlag(f)
			})
		}

		// if this phase has children (not a leaf) it doesn't accept any args directly
		if len(p.Phases) > 0 {
			phaseCmd.Args = cobra.NoArgs
		} else {
			// Use specific validator or inherit from main command
			if p.ArgsValidator != nil {
				phaseCmd.Args = p.ArgsValidator
			} else {
				phaseCmd.Args = cmd.Args // Inherit Args validator from original command
			}
		}

		// adds the command to parent command
		if p.level == 0 {
			phaseCommand.AddCommand(phaseCmd) // Add top-level phases to 'phase'
		} else if p.parent != nil {
			parentCmd, ok := subcommands[p.parent.generatedName]
			if ok {
				parentCmd.AddCommand(phaseCmd) // Add sub-phases to their parent's command
			} else {
				fmt.Printf("[BindToCommand Warning] Parent command for %s not found\n", p.generatedName)
				phaseCommand.AddCommand(phaseCmd) // Fallback: add to top 'phase' cmd
			}
		} else {
			// Should not happen if level > 0 but parent is nil
			phaseCommand.AddCommand(phaseCmd)
		}

		// Store the created command for potential children
		subcommands[p.generatedName] = phaseCmd
		return nil
	})

	// alters the command description to show available phases
	helpText := e.Help(cmd.Use)
	if cmd.Long != "" {
		cmd.Long = fmt.Sprintf("%s\n\n%s\n", cmd.Long, helpText)
	} else {
		cmd.Long = fmt.Sprintf("%s\n\n%s\n", cmd.Short, helpText)
	}

	// adds phase related flags to the main command
	// Use PFlags for persistent flags
	cmd.PersistentFlags().StringSliceVar(&e.Options.SkipPhases, "skip-phases", nil, "List of phases (comma-separated or multiple flags) to be skipped")
}

// --- Simulation of Runner.Run method (Needs to respect Options) ---
func (e *Runner) Run(args []string) error {
	if e.dataInitializer == nil {
		return errors.New("workflow error: DataInitializer not set")
	}
	if e.runCmd == nil {
		// This should be set either by the main command's RunE or the phase subcommand's RunE
		return errors.New("workflow error: runCmd not set internally")
	}

	// Initialize the shared data.
	sharedData, err := e.dataInitializer(e.runCmd, args)
	if err != nil {
		return fmt.Errorf("workflow error: failed to initialize data: %w", err)
	}

	fmt.Printf("--- Starting Workflow Run (Cmd: %s, Filters: %v, Skips: %v) ---\n",
		e.runCmd.Name(), e.Options.FilterPhases, e.Options.SkipPhases)

	// Build skip map for quick lookup
	skipMap := make(map[string]bool)
	for _, skip := range e.Options.SkipPhases {
		// Handle comma-separated values if StringSliceVar wasn't used correctly
		for _, phaseName := range strings.Split(skip, ",") {
			trimmed := strings.TrimSpace(strings.ToLower(phaseName)) // Use lowercase simple name for skip
			if trimmed != "" {
				skipMap[trimmed] = true
			}
		}
	}
	// Build filter map (using generatedName)
	filterMap := make(map[string]bool)
	runFiltered := len(e.Options.FilterPhases) > 0
	if runFiltered {
		for _, filter := range e.Options.FilterPhases {
			filterMap[filter] = true // Filter uses the generatedName
		}
	}

	var executePhase func(*phaseRunner) error
	executePhase = func(p *phaseRunner) error {
		// 1. Check Skips (based on simple name)
		if skipMap[strings.ToLower(p.Name)] {
			fmt.Printf("==> Skipping Phase: %s (due to --skip-phases)\n", p.Name)
			fmt.Println("-------------------------")
			return nil
		}

		// 2. Check Filters (based on generatedName)
		shouldRun := !runFiltered // If no filter, run everything (respecting skips)
		if runFiltered {
			// Check if this phase or any ancestor is in the filter list
			current := p
			for current != nil {
				if filterMap[current.generatedName] {
					shouldRun = true
					break
				}
				current = current.parent
			}
		}

		if !shouldRun {
			// Don't print skip message if just filtering, only if explicitly skipped
			// fmt.Printf("==> Skipping Phase: %s (not in filter)\n", p.Name)
			return nil
		}

		// 3. Execute if not skipped and matches filter (or no filter)
		fmt.Printf("==> Running Phase: %s (%s) / ID: %s\n", p.Name, p.Short, p.generatedName)
		startTime := time.Now()

		// Execute the phase's Run function if it exists and is a leaf or filter targets it directly
		if p.Run != nil && (len(p.Phases) == 0 || filterMap[p.generatedName]) {
			err := p.Run(sharedData)
			if err != nil {
				fmt.Printf("!!! Phase '%s' failed after %v: %v\n", p.Name, time.Since(startTime), err)
				fmt.Println("--- Workflow Run Failed ---")
				return fmt.Errorf("workflow error: phase '%s' failed: %w", p.Name, err)
			}
			fmt.Printf("<== Finished Phase: %s (Duration: %v)\n", p.Name, time.Since(startTime))
		} else if len(p.Phases) > 0 {
			fmt.Printf("    (Phase '%s' has subphases, executing children...)\n", p.Name)
		} else if p.Run == nil {
			fmt.Printf("    (Phase '%s' has no Run function defined)\n", p.Name)
		}

		// 4. Recursively execute children
		for _, child := range p.Phases {
			if err := executePhase(child); err != nil {
				return err // Propagate error up
			}
		}
		fmt.Println("-------------------------") // Separator
		return nil
	}

	// Execute starting from root phases
	if len(e.Phases) == 0 {
		fmt.Println("Workflow has no phases prepared to run.")
	}
	for _, rootPhase := range e.Phases {
		if err := executePhase(rootPhase); err != nil {
			return err // Return the first error encountered
		}
	}

	// Reset filter after run
	e.Options.FilterPhases = nil
	fmt.Println("--- Workflow Run Completed ---")
	return nil
}

// --- End Simulation of the 'workflow' package ---

// --- Demo Specific Code ---

// Concrete RunData
type myWorkflowData struct {
	Log     []string
	SetupOK bool
	AppName string
}

func (d *myWorkflowData) AddLog(msg string) {
	d.Log = append(d.Log, fmt.Sprintf("[%s] %s", time.Now().Format(time.Kitchen), msg))
}

func main() {
	// === 1. Create the Runner ===
	runner := NewRunner()

	// === 2. Define Phases ===
	// Example: Add an additional flag that some phases might inherit
	runner.cmdAdditionalFlags.Bool("dry-run", false, "Perform a dry run without making changes")

	// Phase: Setup
	setupPhase := Phase{
		Name:  "Setup",
		Short: "Initial setup phase",
		Long:  "Performs initial system setup, creates directories, etc.",
		Run: func(data RunData) error {
			d := data.(*myWorkflowData) // Type assertion
			d.AddLog("Setup phase running.")
			fmt.Println("   Running Setup...")
			// Access inherited flag (example)
			dryRun, _ := runner.runCmd.Flags().GetBool("dry-run")
			if dryRun {
				d.AddLog("Setup: Dry run enabled.")
				fmt.Println("   (Dry Run) Would create directories...")
			} else {
				fmt.Println("   Creating directories...")
				time.Sleep(50 * time.Millisecond)
			}
			d.SetupOK = true
			d.AddLog("Setup phase finished.")
			return nil
		},
		// Sub-phases for Setup
		SubPhases: []Phase{
			{
				Name:  "Download",
				Short: "Download necessary assets",
				Run: func(data RunData) error {
					d := data.(*myWorkflowData)
					d.AddLog("Download sub-phase running.")
					fmt.Println("      Downloading assets...")
					time.Sleep(80 * time.Millisecond)
					d.AddLog("Download sub-phase finished.")
					return nil
				},
			},
			{
				Name:   "Validate",
				Short:  "Validate downloaded assets",
				Hidden: true, // Example of a hidden phase
				Run: func(data RunData) error {
					d := data.(*myWorkflowData)
					d.AddLog("Validate sub-phase running.")
					fmt.Println("      Validating assets (hidden)...")
					time.Sleep(40 * time.Millisecond)
					d.AddLog("Validate sub-phase finished.")
					return nil
				},
			},
		},
	}

	// Phase: Build
	buildPhase := Phase{
		Name:    "Build",
		Short:   "Build the application",
		Example: `myapp build --target=linux/amd64`, // Example usage
		Run: func(data RunData) error {
			d := data.(*myWorkflowData)
			d.AddLog("Build phase running.")
			fmt.Println("   Running Build...")
			if !d.SetupOK {
				return errors.New("build error: Setup was not successful")
			}
			target, _ := runner.runCmd.Flags().GetString("target") // Access local flag
			fmt.Printf("   Building for target: %s\n", target)
			time.Sleep(100 * time.Millisecond)
			d.AddLog(fmt.Sprintf("Build phase finished (target: %s).", target))
			return nil
		},
		InheritFlags:  []string{"verbose"},                                    // Only inherit 'verbose' from root
		LocalFlags:    pflag.NewFlagSet("build-flags", pflag.ContinueOnError), // Phase-specific flags
		ArgsValidator: cobra.NoArgs,                                           // Build doesn't take positional args
	}
	// Add local flag to build phase
	buildPhase.LocalFlags.String("target", "linux/amd64", "Build target platform")

	// Phase: Deploy
	deployPhase := Phase{
		Name:    "Deploy",
		Short:   "Deploy the application",
		Aliases: []string{"push"},
		Run: func(data RunData) error {
			d := data.(*myWorkflowData)
			d.AddLog("Deploy phase running.")
			fmt.Println("   Running Deploy...")
			server, _ := runner.runCmd.Flags().GetString("server")
			fmt.Printf("   Deploying app '%s' to server: %s\n", d.AppName, server)
			time.Sleep(120 * time.Millisecond)
			d.AddLog(fmt.Sprintf("Deploy phase finished (server: %s).", server))
			return nil
		},
		LocalFlags: pflag.NewFlagSet("deploy-flags", pflag.ContinueOnError),
	}
	deployPhase.LocalFlags.String("server", "production.example.com", "Target deployment server")

	// === 3. Set Data Initializer ===
	runner.SetDataInitializer(func(cmd *cobra.Command, args []string) (RunData, error) {
		fmt.Println("[Data Initializer Called]")
		// Get flag value potentially set on the command triggering init
		appName, _ := cmd.Flags().GetString("app-name")
		initialData := &myWorkflowData{
			Log:     make([]string, 0),
			AppName: appName, // Use flag value
		}
		initialData.AddLog("Data object created by Initializer.")
		return initialData, nil
	})

	// === 4. Create Root Cobra Command ===
	var rootCmd = &cobra.Command{
		Use:   "myapp",
		Short: "A demonstration application with workflow phases",
		Long:  "This application demonstrates how to use workflow phases integrated with Cobra.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			fmt.Println("--- Root PersistentPreRun ---")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is called when 'myapp' is run without subcommands or 'phase'
			fmt.Printf("[Root RunE triggered for '%s']\n", cmd.Name())
			// If the root command is called directly, run the whole workflow
			runner.runCmd = cmd               // Set the command triggering the full run
			runner.Options.FilterPhases = nil // Ensure no filter is active
			// Skip phases will be parsed from cmd.PersistentFlags() by BindToCommand
			return runner.Run(args)
		},
	}
	// Add flags to the root command
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().String("app-name", "DefaultApp", "Name of the application being processed")

	// === 5. Add Phases to Runner ===
	runner.AppendPhase(setupPhase)
	runner.AppendPhase(buildPhase)
	runner.AppendPhase(deployPhase)

	// === 6. *** Bind Runner to Command *** ===
	// This MUST be called AFTER adding phases and setting up the root command flags
	runner.BindToCommand(rootCmd)

	// === 7. Execute ===
	fmt.Println("--- Executing Cobra Command ---")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("--- Cobra Command Execution Finished ---")
}
