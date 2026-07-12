package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/graphql-go/graphql"
)

// Job struct
type Job struct {
	ID             int      `json:"id"`
	Position       string   `json:"position"`
	Company        string   `json:"company"`
	Description    string   `json:"description"`
	SkillsRequired []string `json:"skillsRequired"`
	Location       string   `json:"location"`
	EmploymentType string   `json:"employmentType"`
}

var jobType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Job",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"position": &graphql.Field{
				Type: graphql.String,
			},
			"company": &graphql.Field{
				Type: graphql.String,
			},
			"description": &graphql.Field{
				Type: graphql.String,
			},
			"location": &graphql.Field{
				Type: graphql.String,
			},
			"employmentType": &graphql.Field{
				Type: graphql.String,
			},
			"skillsRequired": &graphql.Field{
				Type: graphql.NewList(graphql.String),
			},
		},
	},
)

type reqBody struct {
	Query string `json:"query"`
}

type jobLoader func() ([]Job, error)

func gqlHandler() http.Handler {
	return gqlHandlerWithLoader(retrieveJobsFromFile("data.json"))
}

func gqlHandlerWithLoader(loadJobs jobLoader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "No query data", 400)
			return
		}
		defer r.Body.Close()

		var rBody reqBody
		err := json.NewDecoder(r.Body).Decode(&rBody)
		if err != nil {
			http.Error(w, "Error parsing JSON request body", 400)
			return
		}

		result, err := processQuery(rBody.Query, loadJobs)
		if err != nil {
			http.Error(w, "Error processing GraphQL query", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(result))
	})
}

func processQuery(query string, loadJobs jobLoader) (string, error) {
	if loadJobs == nil {
		return "", errors.New("job loader is required")
	}
	schema, err := gqlSchema(loadJobs)
	if err != nil {
		return "", err
	}
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	rJSON, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(rJSON), nil
}

// retrieveJobsFromFile opens a JSON data file and returns the decoded jobs.
func retrieveJobsFromFile(path string) jobLoader {
	return func() ([]Job, error) {
		jsonDataFromFile, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var jobsData []Job
		err = json.Unmarshal(jsonDataFromFile, &jobsData)
		if err != nil {
			return nil, err
		}

		return jobsData, nil
	}
}

// Define the GraphQL Schema
func gqlSchema(queryJobs jobLoader) (graphql.Schema, error) {
	fields := graphql.Fields{
		"jobs": &graphql.Field{
			Type:        graphql.NewList(jobType),
			Description: "All Jobs",
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return queryJobs()
			},
		},
		"job": &graphql.Field{
			Type:        jobType,
			Description: "Get Jobs by ID",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, success := params.Args["id"].(int)
				if success {
					jobs, err := queryJobs()
					if err != nil {
						return nil, err
					}
					for _, job := range jobs {
						if job.ID == id {
							return job, nil
						}
					}
				}
				return nil, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return graphql.Schema{}, err
	}

	return schema, nil
}
