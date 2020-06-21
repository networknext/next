
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func dataflow(env Environment) {

	pipelines, err := loadPipelines()
	if err != nil {
		// todo: print error
		return
	}

	pipelineStates, err := getExistingPipelineStates()
	if err != nil {
		// todo: print error
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Class Name", "Status", "Last Change (UTC)"})

	for _, pipeline := range pipelines {
		pipelineState, ok := pipelineStates[fmt.Sprintf("%s-pipeline", pipeline.Name)]
		status := "Not Deployed"
		lastChange := ""
		if ok {
			lastChange = pipelineState.StateTime
			if pipelineState.State == "Running" {
				status = aurora.BrightGreen("Running").String()
			} else {
				status = aurora.BrightYellow(pipelineState.State).String()
			}
		}
		table.Append([]string{pipeline.Name, pipeline.ClassName, status, lastChange})
	}

	// todo: update this to the nicer table we use
	table.Render()
}

type pipelineDefinition struct {
	Name                  string                       `json:"name"`
	ClassName             string                       `json:"className"`
	MaxWorkers            int                          `json:"maxWorkers"`
	DesiredWorkers        int                          `json:"desiredWorkers"`
	WorkerSize            string                       `json:"workerSize"`
	Args                  map[string]string            `json:"args"`
	EnvArgs               map[string]map[string]string `json:"envArgs"`
	EnableStreamingEngine bool                         `json:"enableStreamingEngine"`
}

type pipelineEnvironment struct {
	Project string
	Env     string
	Dataset string
}

type dataflowPipelineStatus struct {
	CreationTime string `json:"creationTime"`
	ID           string `json:"id"`
	Location     string `json:"location"`
	Name         string `json:"name"`
	State        string `json:"state"`
	StateTime    string `json:"stateTime"`
	Type         string `json:"type"`
}

func getExistingPipelineStates() (map[string]*dataflowPipelineStatus, error) {
	ok, output := runCommandQuiet(
		"gcloud",
		[]string{
			fmt.Sprintf("--project=%s", getProject()),
			"dataflow",
			"jobs",
			"list",
			"--format",
			"json",
			"--region",
			"us-central1",
		},
		true,
	)
	if !ok {
		return nil, fmt.Errorf("unable to retrieve Dataflow jobs")
	}

	var states []*dataflowPipelineStatus

	err := json.Unmarshal([]byte(output), &states)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*dataflowPipelineStatus)

	for _, state := range states {
		if state.State == "Running" || state.State == "Stopped" || state.State == "Draining" || state.State == "Pending" || state.State == "Cancelling" || state.State == "Queued" {
			results[state.Name] = state
		}
	}

	return results, nil
}

func getProject() string {
	// todo: depending on env
	return "network-next-v3-prod"
	/*
	if environment == "v3" {
		return "network-next-v3-prod"
	} else if environment == "v3-dev" {
		return "network-next-v3-dev"
	} else if environment == "local" {
		return "network-next-local"
	} else {
		panic("unknown environment (must be v3, v3-dev or local)")
	}
	*/
}

func getMachineType(workerSize string) string {
	return "n1-standard-32"
	/*
	if environment == "v3" {
		return workerSize // increase instance size, as recommended by Google Support
	}
	return "n1-standard-4"
	*/
}

func getMaxWorkers(pipelineMax int) int {
	return 16
	/*
	if pipelineMax == 0 {
		panic("maxWorkers must be set")
	}

	if environment == "v3" {
		return pipelineMax
	}

	return 2
	*/
}

func getDesiredWorkers(pipelineDesired int, pipelineMax int) int {
	return 16
	/*
	if environment == "v3" {
		if pipelineMax < pipelineDesired {
			return pipelineMax
		}

		return pipelineDesired
	}

	return 2
	*/
}

func getPipelineEnvironment(pipelineName string) pipelineEnvironment {
	return pipelineEnvironment{
		Project: "network-next-v3-prod",
		Env:     "v3",
		Dataset: "v3",
	}
	/*
	if environment == "v3" {
		return pipelineEnvironment{
			Project: "network-next-v3-prod",
			Env:     "v3",
			Dataset: "v3",
		}
	} else if environment == "v3-dev" {
		return pipelineEnvironment{
			Project: "network-next-v3-dev",
			Env:     "v3-dev",
			Dataset: "v3_dev",
		}
	} else if environment == "local" {
		return pipelineEnvironment{
			Project: "network-next-local",
			Env:     "local",
			Dataset: "local",
		}
	} else {
		panic("unknown environment (must be v3, v3-dev or local)")
	}
	*/
}

func loadPipelines() (map[string]*pipelineDefinition, error) {
	var rawPipelines []pipelineDefinition

	rawJSON, err := ioutil.ReadFile("dataflow/pipelines.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawJSON, &rawPipelines)
	if err != nil {
		return nil, err
	}

	pipelines := make(map[string]*pipelineDefinition)
	for _, pipeline := range rawPipelines {
		processedArgs := make(map[string]string)
		pipelineEnv := getPipelineEnvironment(pipeline.Name)

		if pipeline.Args != nil {
			for k, v := range pipeline.Args {
				kk := fmt.Sprintf("%s-%s", pipeline.Name, k)
				tmpl := template.Must(template.New(kk).Parse(v))
				var tpl bytes.Buffer
				err := tmpl.Execute(&tpl, pipelineEnv)
				if err != nil {
					return nil, err
				}
				processedArgs[k] = string(tpl.Bytes())
			}
		}

		if pipeline.EnvArgs != nil {
			// envArgs := pipeline.EnvArgs[environment]
			envArgs := pipeline.EnvArgs["v3"]
			if envArgs != nil {
				for k, v := range envArgs {
					kk := fmt.Sprintf("%s-%s", pipeline.Name, k)
					tmpl := template.Must(template.New(kk).Parse(v))
					var tpl bytes.Buffer
					err := tmpl.Execute(&tpl, pipelineEnv)
					if err != nil {
						return nil, err
					}
					processedArgs[k] = string(tpl.Bytes())
				}
			}
		}

		// also change this
		pipelines[pipeline.Name] = &pipelineDefinition{
			Name:                  pipeline.Name,
			ClassName:             pipeline.ClassName,
			MaxWorkers:            pipeline.MaxWorkers,
			DesiredWorkers:        pipeline.DesiredWorkers,
			WorkerSize:            pipeline.WorkerSize,
			EnableStreamingEngine: pipeline.EnableStreamingEngine,
			Args:                  processedArgs,
		}
	}

	return pipelines, nil
}

func filterPipelines(c *cli.Context, pipelines map[string]*pipelineDefinition, needOne bool) ([]*pipelineDefinition, error) {

	result := make([]*pipelineDefinition, 0)

	if needOne {
		if c.String("pipeline") != "" {
			pipeline := pipelines[c.String("pipeline")]
			if pipeline == nil {
				return nil, fmt.Errorf("can't find %s pipeline in pipelines.json", c.String("pipeline"))
			}
			result = append(result, pipeline)
		}
	} else {
		pipeline := pipelines["platform"]
		if pipeline == nil {
			return nil, fmt.Errorf("can't find platform pipeline in pipelines.json")
		}
		result = append(result, pipeline)

		pipeline = pipelines["export"]
		if pipeline == nil {
			return nil, fmt.Errorf("can't find export pipeline in pipelines.json")
		}
		result = append(result, pipeline)
	}

	return result, nil

}

/*


var environment string

func main() {

	app := cli.NewApp()

	app.Name = "next dataflow"
	app.HelpName = "next dataflow"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "env",
			Value:       "v3-dev",
			Usage:       "environment to deploy Dataflow jobs in (one of: v3, v3-dev, local; default: v3-dev)",
			Destination: &environment,
		},
	}
	app.Usage = "Tools for deploying and executing Dataflow pipelines"
	app.HideVersion = true

	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "List the Dataflow pipeline and it's status",
			Action: func(c *cli.Context) error {

				pipelines, err := loadPipelines()
				if err != nil {
					return err
				}

				pipelineStates, err := getExistingPipelineStates()
				if err != nil {
					return err
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Name", "Class Name", "Status", "Last Change (UTC)"})

				for _, pipeline := range pipelines {
					pipelineState, ok := pipelineStates[fmt.Sprintf("%s-pipeline", pipeline.Name)]
					status := "Not Deployed"
					lastChange := ""
					if ok {
						lastChange = pipelineState.StateTime
						if pipelineState.State == "Running" {
							status = aurora.BrightGreen("Running").String()
						} else {
							status = aurora.BrightYellow(pipelineState.State).String()
						}
					}
					table.Append([]string{pipeline.Name, pipeline.ClassName, status, lastChange})
				}

				table.Render()

				return nil

			},
		},
		{
			Name:  "deploy",
			Usage: "Deploy the Dataflow pipeline",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{Name: "remove"},
				&cli.StringSliceFlag{Name: "rename"},
				&cli.StringFlag{
					Name:  "pipeline, p",
					Value: "",
					Usage: "The pipeline to deploy (if none specified, deploys all).",
				},
			},
			Action: func(c *cli.Context) error {

				pipelines, err := loadPipelines()
				if err != nil {
					return err
				}

				pipelineStates, err := getExistingPipelineStates()
				if err != nil {
					return err
				}

				os.Chdir("dataflow")

				pipelineSlice, err := filterPipelines(c, pipelines, c.String("pipeline") != "")
				if err != nil {
					return err
				}

				transformMap := make(map[string]string)
				for _, e := range c.StringSlice("remove") {
					transformMap[e] = ""
				}
				for _, e := range c.StringSlice("rename") {
					c := strings.SplitN(e, "=", 2)
					if len(c) == 2 {
						transformMap[c[0]] = c[1]
					}
				}

				for _, pipeline := range pipelineSlice {
					altSuffix := ""
					altStatus, altOk := pipelineStates[fmt.Sprintf("%s-pipeline-alt", pipeline.Name)]
					status, ok := pipelineStates[fmt.Sprintf("%s-pipeline", pipeline.Name)]
					if altOk {
						altSuffix = "-alt"
						if altStatus.State == "Draining" || altStatus.State == "Cancelling" {
							altSuffix = ""
						}
					} else if ok {
						if status.State == "Draining" || status.State == "Cancelling" {
							altSuffix = "-alt"
						}
					}

					args := []string{
						"--runner=DataflowRunner",
						"--streaming=true",
						"--autoscalingAlgorithm=THROUGHPUT_BASED",
						"--profilingAgentConfiguration={\"APICurated\":true,\"NativeStacks\":true}",
						// do not --enableStreamingEngine, it is expensive with no real benefit
						"--zone=us-central1-a",
						"--region=us-central1",
						fmt.Sprintf(
							"--workerDiskType=compute.googleapis.com/projects/%s/zones/us-central1-a/diskTypes/pd-ssd",
							getProject(),
						),
						fmt.Sprintf("--numWorkers=%d", getDesiredWorkers(pipeline.DesiredWorkers, pipeline.MaxWorkers)),
						fmt.Sprintf("--workerMachineType=%s", getMachineType(pipeline.WorkerSize)),
						fmt.Sprintf("--serviceAccount=backend@%s.iam.gserviceaccount.com", getProject()),
						fmt.Sprintf("--project=%s", getProject()),
						fmt.Sprintf("--jobName=%s-pipeline%s", pipeline.Name, altSuffix),
					}
					if pipeline.EnableStreamingEngine {
						args = append(
							args,
							"--enableStreamingEngine",
							"--diskSizeGb=30",
						)
					} else {
						args = append(args, fmt.Sprintf("--maxNumWorkers=%d", getMaxWorkers(pipeline.MaxWorkers)))
					}

					for k, v := range pipeline.Args {
						args = append(args, fmt.Sprintf("--%s=%s", k, v))
					}
					if len(transformMap) > 0 {
						b, _ := json.Marshal(transformMap)
						args = append(args, fmt.Sprintf("--transformNameMapping=%s", string(b)))
					}

					status, ok = pipelineStates[fmt.Sprintf("%s-pipeline%s", pipeline.Name, altSuffix)]
					if ok && status.State != "Draining" && status.State != "Cancelling" {
						args = append(args, "--update")
					}

					// Pass arguments through a file, since on Windows there's no safe way to pass
					// JSON through multiple levels of indirection on the command line.
					f, err := ioutil.TempFile("", "args")
					if err != nil {
						return err
					}
					f.Write([]byte(strings.Join(args, "\n")))
					f.Close()
					defer os.Remove(f.Name())
					os.Setenv("PIPELINE_ARGUMENTS_FILE", f.Name())

					if !common.RunCommand("mvn", []string{
						"compile",
						"-B",
						"exec:java",
						fmt.Sprintf("-Dexec.mainClass=com.networknext.dataflow.%s", pipeline.ClassName),
					}) {
						return fmt.Errorf("mvn compile failed")
					}
				}

				return nil

			},
		},
		{
			Name:  "run",
			Usage: "Runs the Dataflow pipeline in your local environment",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "pipeline, p",
					Value: "platform",
					Usage: "The pipeline to run (defaults to 'platform').",
				},
			},
			Action: func(c *cli.Context) error {

				// force local environment
				environment = "local"

				pipelines, err := loadPipelines()
				if err != nil {
					return err
				}

				os.Chdir("dataflow")

				pipelineSlice, err := filterPipelines(c, pipelines, true)
				if err != nil {
					return err
				}

				if len(pipelineSlice) != 1 {
					return fmt.Errorf("you must name exactly one pipeline")
				}

				pipeline := pipelineSlice[0]

				args := []string{
					"--runner=DirectRunner",
					"--streaming=true",
					"--targetParallelism=16",
					"--blockOnRun=true",
					fmt.Sprintf("--project=%s", getProject()),
				}
				for k, v := range pipeline.Args {
					args = append(args, fmt.Sprintf("--%s=%s", k, v))
				}

				wd, _ := os.Getwd()
				gac := fmt.Sprintf("%s/../services/common/creds/network-next-local.json", wd)

				os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8493")
				os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:8491")
				os.Setenv("BIGTABLE_EMULATOR_HOST", "localhost:8490")
				os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8492")
				os.Setenv("REDIS_HOST", "localhost:6380")
				os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gac)
				os.Setenv("DIRECT_RUNNER", "1")

				if !common.RunCommand("mvn", []string{
					"compile",
					"-B",
					"exec:java",
					fmt.Sprintf("-Dexec.mainClass=com.networknext.dataflow.%s", pipeline.ClassName),
					"-Dexec.args=" + strings.Join(args, " "),
				}) {
					return fmt.Errorf("mvn compile failed")
				}

				return nil

			},
		},
	}

	os.Chdir("../..")

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
*/
