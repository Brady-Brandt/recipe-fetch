package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)


func GetRecipeJson(url string) (map[string]any, error) {	
	var recipeJson map[string]any

	client := &http.Client{}	
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create http request")
		return recipeJson, err
	}

	// if this isn't set we may be denied access to the website
	req.Header.Set("User-Agent", "Linux x86_64")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to send http request")
		return recipeJson, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	var jsonStart = strings.Index(html, "application/ld+json")

	if jsonStart == -1 {
		return recipeJson, errors.New("Could not Find application/ld+json")
	}

	for c := html[jsonStart]; c != '{'; {
		jsonStart += 1
		c = html[jsonStart]
	}

	var jsonEnd = jsonStart + strings.Index(html[jsonStart:], "</script>")

	err = json.Unmarshal([]byte(html[jsonStart:jsonEnd]), &recipeJson)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse json")
	}

	return recipeJson, err
}


func GetIngredients(jsonField map[string]any, ingredients *[]any) {
	_, hasIngredients:= jsonField["recipeIngredient"]
	if hasIngredients {
		*ingredients = jsonField["recipeIngredient"].([]any)
	}
}


func GetInstructions(jsonField map[string]any, instructions*[]any) {
	_, hasInstructions:= jsonField["recipeInstructions"]
	if hasInstructions {
		*instructions = jsonField["recipeInstructions"].([]any)
	}
}


func GetInstruction(instrMap map[string]any) (string, error) {
	var instr string = ""
	if step, hasKey := instrMap["name"]; hasKey {
		instr += step.(string) + " "
	}
	if step, hasKey := instrMap["text"]; hasKey {
		// on some websites these 2 fields are the same
		// only want to capture both when they are different
		if instr != step.(string) + " "{
			instr += step.(string)
		}
	}

	if len(instr) != 0 {
		return instr, nil
	} else{
		return "", errors.New("Could not find directions")
	}
}

func ConstructMD(recipeName string, ingredients []any, instructions []any) error {
	fileName := recipeName + ".md"
	f, err := os.Create(fileName)

	if err != nil {
		return err
	}

	defer f.Close()
	f.WriteString("# " + recipeName + "\n\n")

	f.WriteString("## Ingredients\n")

	for _, ingredient := range ingredients {
		ingredient := ingredient.(string)
		if ingredient == strings.ToUpper(ingredient) {
			f.WriteString("\n### " + ingredient + "\n\n")
		} else{
			f.WriteString("- " + ingredient + "\n")
		}
	}

	f.WriteString("\n## Instructions\n")

	for step_cnt, step := range instructions {
		instr, err := GetInstruction(step.(map[string]any))
		if err != nil {
			return err
		}
		f.WriteString("### Step " + strconv.Itoa(step_cnt + 1) + "\n")
		f.WriteString(instr + "\n")
	}
	fmt.Println("Sucessfully Created " + fileName)
	return nil
}

func ConstructHtml(recipeName string, ingredients []any, instructions []any) error {
	fileName := recipeName + ".html"
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer f.Close()
	fmt.Fprintf(f,
`<!DOCTYPE html>
<html>
<head>
	<title>%s</title>
</head>
<body>
<h1>%s</h1>`, recipeName, recipeName)


	f.WriteString("\n<h2>Ingredients</h2>\n")
	f.WriteString("<ul>\n")
	for _, ingredient := range ingredients {
		ingredient := ingredient.(string)
		if ingredient == strings.ToUpper(ingredient) {
			fmt.Fprintf(f, "<h3> %s </h3>\n", ingredient)
		} else{
			fmt.Fprintf(f, " <li>%s</li>\n", ingredient)
		}
	}

	f.WriteString("</ul>\n")

	f.WriteString("<h2>Instructions</h2>\n")
	f.WriteString("<ol>\n")

	for _, step := range instructions {
		instr, err := GetInstruction(step.(map[string]any))
		if err != nil {
			return err
		}
		fmt.Fprintf(f, "<li><p>%s</p></li>\n", instr)
	}

	f.WriteString("</lo>\n")
	f.WriteString("</body>\n")
	f.WriteString("</html>\n")
	fmt.Println("Sucessfully Created " + fileName)
	return nil
}

func PrintUsage(){
	fmt.Fprintln(os.Stderr, "Usage: ")
	flag.PrintDefaults()	
	os.Exit(1)
}


func main(){
	var url, recipeName string 
	var generateHtml bool = false

	flag.StringVar(&url, "url", "", "URL of the website")
	flag.StringVar(&recipeName, "o", "", "Name of the recipe")
	flag.BoolVar(&generateHtml, "html", false, "Generates HTML instead of markdown")

	flag.Parse()

	if len(os.Args) < 2 {
		PrintUsage()
	}


	if len(url) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Missing URL")
		os.Exit(1)
	}

	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Error excess arguements")
		PrintUsage()
	}


	// if no recipeName is inputted use the last part of the url as the recipeName
	if len(recipeName) == 0 {
		splitUrl := strings.Split(url, "/")
		for i := len(splitUrl) - 1; i > 0; i-- {
			if len(splitUrl[i]) != 0 {
				recipeName = splitUrl[i]
				break
			}
		}
	}

	jsonRecipe, err := GetRecipeJson(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	

	var ingredients  []any = nil
	var instructions []any = nil

	if graph, hasKey := jsonRecipe["@graph"]; hasKey {
		for _, item := range graph.([]any) {
			field := item.(map[string]any)
			GetIngredients(field, &ingredients)	
			GetInstructions(field, &instructions)
		}
	} else{
		GetIngredients(jsonRecipe, &ingredients)
		GetInstructions(jsonRecipe, &instructions)
	}

	if ingredients == nil {
		fmt.Fprintln(os.Stderr, "Failed to get Ingredients")
		os.Exit(1)

	}

	if instructions == nil {
		fmt.Fprintln(os.Stderr, "Failed to get Instructions")
		os.Exit(1)
	}


	if generateHtml {
		err = ConstructHtml(recipeName, ingredients, instructions)
	} else{
		err = ConstructMD(recipeName, ingredients, instructions)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
