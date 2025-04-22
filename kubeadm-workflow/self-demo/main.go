package main

import (
	"errors"
	"fmt"
	"time"
	// We simulate cobra.Command with nil, so we don't need the actual import
	// "github.com/spf13/cobra"
	// Instead, define a type alias if needed for the signature, or just use interface{}
	// For simplicity, the simulated SetDataInitializer won't actually use the command object.
	// type Command = interface{} // Dummy type if we needed strict signature matching
)

// --- Start Simulation of the 'workflow' package ---

// RunData is the interface type for data shared between phases.
// Using interface{} as the simplest form based on type assertion in the example.
type RunData interface{}

// Phase represents a single step in the workflow.
type Phase struct {
	Name  string
	Short string
	Run   func(data RunData) error // The function to execute for this phase
}

// Runner manages and executes a sequence of phases.
type Runner struct {
	phases          []Phase
	dataInitializer func(cmd /* *cobra.Command */ interface{}, args []string) (RunData, error)
}

// NewRunner creates a new workflow Runner.
func NewRunner() *Runner {
	return &Runner{
		phases: make([]Phase, 0),
	}
}

// AppendPhase adds a phase to the end of the workflow sequence.
func (e *Runner) AppendPhase(p Phase) {
	e.phases = append(e.phases, p)
}

// SetDataInitializer sets the function used to create the initial RunData.
// We use 'interface{}' to represent the *cobra.Command argument for this simulation.
func (e *Runner) SetDataInitializer(initializer func(cmd interface{}, args []string) (RunData, error)) {
	e.dataInitializer = initializer
}

// Run executes the workflow phases sequentially.
func (e *Runner) Run(args []string) error {
	if e.dataInitializer == nil {
		return errors.New("workflow error: DataInitializer not set")
	}

	// Initialize the shared data using the provided function.
	// Pass nil for the simulated cobra command.
	sharedData, err := e.dataInitializer(nil, args)
	if err != nil {
		return fmt.Errorf("workflow error: failed to initialize data: %w", err)
	}

	fmt.Println("--- Starting Workflow ---")
	if len(e.phases) == 0 {
		fmt.Println("Workflow has no phases to run.")
		return nil
	}

	// Execute each phase in order
	for _, phase := range e.phases {
		fmt.Printf("==> Running Phase: %s (%s)\n", phase.Name, phase.Short)
		startTime := time.Now()

		// Execute the phase's Run function, passing the shared data
		err := phase.Run(sharedData)
		if err != nil {
			// Stop workflow execution on the first error
			fmt.Printf("!!! Phase '%s' failed after %v: %v\n", phase.Name, time.Since(startTime), err)
			fmt.Println("--- Workflow Failed ---")
			return fmt.Errorf("workflow error: phase '%s' failed: %w", phase.Name, err)
		}

		fmt.Printf("<== Finished Phase: %s (Duration: %v)\n", phase.Name, time.Since(startTime))
		fmt.Println("-------------------------") // Separator between phases
	}

	fmt.Println("--- Workflow Completed Successfully ---")
	return nil
}

// --- End Simulation of the 'workflow' package ---

// --- Demo Specific Code ---

// myWorkflowData is our concrete implementation for RunData for this demo.
// Using a pointer so modifications in one phase are seen by the next.
type myWorkflowData struct {
	ActionLog []string
	Counter   int
}

// (Optional) A helper method for our specific data type
func (d *myWorkflowData) AddLog(message string) {
	d.ActionLog = append(d.ActionLog, fmt.Sprintf("[%s] %s", time.Now().Format(time.Kitchen), message))
}

func main() {
	// 1. Define the phases
	phaseInitialize := Phase{
		Name:  "Initialize System",
		Short: "Sets up the initial state.",
		Run: func(data RunData) error {
			// Type assert the RunData to our specific type
			d, ok := data.(*myWorkflowData)
			if !ok {
				// This should not happen if DataInitializer returns the correct type
				return errors.New("invalid RunData type received in phase 'Initialize'")
			}

			fmt.Println("   Initializing...")
			d.AddLog("System initialization started.")
			d.Counter = 0 // Set initial value
			fmt.Printf("   Initial counter: %d\n", d.Counter)
			time.Sleep(50 * time.Millisecond) // Simulate work
			d.AddLog("System initialization complete.")
			return nil
		},
	}

	phaseProcess := Phase{
		Name:  "Process Data",
		Short: "Performs the main processing.",
		Run: func(data RunData) error {
			d, ok := data.(*myWorkflowData)
			if !ok {
				return errors.New("invalid RunData type received in phase 'Process'")
			}

			fmt.Println("   Processing...")
			d.AddLog("Processing data.")
			d.Counter += 10 // Modify the data
			fmt.Printf("   Counter incremented to: %d\n", d.Counter)
			time.Sleep(100 * time.Millisecond) // Simulate work

			// Example of a potential failure point
			// if d.Counter > 5 {
			//  d.AddLog("ERROR: Counter exceeded threshold during processing!")
			// 	return errors.New("counter threshold exceeded")
			// }

			d.AddLog("Processing complete.")
			return nil
		},
	}

	phaseFinalize := Phase{
		Name:  "Finalize System",
		Short: "Cleans up and reports final state.",
		Run: func(data RunData) error {
			d, ok := data.(*myWorkflowData)
			if !ok {
				return errors.New("invalid RunData type received in phase 'Finalize'")
			}

			fmt.Println("   Finalizing...")
			d.AddLog("Finalization started.")
			fmt.Printf("   Final counter value: %d\n", d.Counter)
			fmt.Println("   Action Log:")
			for _, entry := range d.ActionLog {
				fmt.Printf("     - %s\n", entry)
			}
			time.Sleep(30 * time.Millisecond) // Simulate work
			d.AddLog("Finalization complete.")
			return nil
		},
	}

	// 2. Create a Runner instance
	myWorkflowRunner := NewRunner()

	// 3. Set the Data Initializer
	myWorkflowRunner.SetDataInitializer(func(cmd interface{}, args []string) (RunData, error) {
		fmt.Println("[Data Initializer Called]")
		// Create the initial data structure that will be passed to the first phase
		initialData := &myWorkflowData{
			ActionLog: make([]string, 0),
			Counter:   -1, // Start with a value to see it change
		}
		initialData.AddLog("Data object created by Initializer.")
		return initialData, nil // Return the pointer to our data struct
	})

	// 4. Append phases in the desired execution order
	myWorkflowRunner.AppendPhase(phaseInitialize)
	myWorkflowRunner.AppendPhase(phaseProcess)
	myWorkflowRunner.AppendPhase(phaseFinalize)

	// 5. Run the workflow
	// We pass an empty slice for args as this demo doesn't use command-line args
	err := myWorkflowRunner.Run([]string{})
	if err != nil {
		fmt.Printf("\n*** Workflow execution failed: %v\n", err)
	} else {
		fmt.Println("\n*** Workflow finished successfully.")
	}
}
