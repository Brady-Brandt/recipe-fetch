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

type Recipe struct {
	Ingredients  []any
	Instructions []any
	CookTime     string
	PrepTime     string
	TotalTime    string
}


func (recipe *Recipe) GetIngredients(jsonField map[string]any){
	_, hasIngredients:= jsonField["recipeIngredient"]
	if hasIngredients {
		recipe.Ingredients = jsonField["recipeIngredient"].([]any)
	}
}

func (recipe *Recipe) GetInstructions(jsonField map[string]any){
	_, hasInstructions:= jsonField["recipeInstructions"]
	if hasInstructions {
		recipe.Instructions = jsonField["recipeInstructions"].([]any)
	}
}

func (recipe *Recipe) GetCookTime(jsonField map[string]any){
	_, hasCookTime:= jsonField["cookTime"]
	if hasCookTime {
		recipe.CookTime = FormatTime(jsonField["cookTime"].(string))
	}
}

func (recipe *Recipe) GetPrepTime(jsonField map[string]any){
	_, hasPrepTime:= jsonField["prepTime"]
	if hasPrepTime {
		recipe.PrepTime = FormatTime(jsonField["prepTime"].(string))
	}
}

func (recipe *Recipe) GetTotalTime(jsonField map[string]any){
	_, hasTotalTime:= jsonField["totalTime"]
	if hasTotalTime {
		recipe.TotalTime = FormatTime(jsonField["totalTime"].(string))
	}
}

func FormatTime(time string) string {
	time = strings.TrimPrefix(time, "PT")
	time = strings.TrimPrefix(time, "0H")
	if strings.HasPrefix(time, "1H") {
		time = strings.Replace(time, "H", " hour ", 1)
	} else{
		time = strings.Replace(time, "H", " hours ", 1)
	}
	time = strings.Replace(time, "M", " minutes ", 1)
	return time
}

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

	for {
		c := html[jsonEnd];
		if c == '}' {
			jsonEnd += 1
			break
		}
		jsonEnd -= 1
		c = html[jsonEnd]
	}

	err = json.Unmarshal([]byte(html[jsonStart:jsonEnd]), &recipeJson)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse json")
	}

	return recipeJson, err
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

func ConstructMD(url string, recipeName string, recipe Recipe) error {
	fileName := recipeName + ".md"
	f, err := os.Create(fileName)

	if err != nil {
		return err
	}

	defer f.Close()
	f.WriteString("# " + recipeName + "\n\n")

	fmt.Fprintf(f, "### [Orignal Recipe](%s)\n", url)

	if(len(recipe.PrepTime) != 0){
		f.WriteString("### Prep Time: " + recipe.PrepTime + "\n")
	}

	if(len(recipe.CookTime) != 0){
		f.WriteString("### Cook Time: " + recipe.CookTime + "\n")
	}

	if(len(recipe.TotalTime) != 0){
		f.WriteString("### Total Time: " + recipe.TotalTime + "\n")
	}

	f.WriteString("## Ingredients\n")

	for _, ingredient := range recipe.Ingredients {
		ingredient := ingredient.(string)
		if ingredient == strings.ToUpper(ingredient) {
			f.WriteString("\n### " + ingredient + "\n\n")
		} else{
			f.WriteString("- " + ingredient + "\n")
		}
	}

	f.WriteString("\n## Instructions\n")

	for step_cnt, step := range recipe.Instructions {
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

func ConstructHtml(url string, recipeName string, recipe Recipe) error {
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
<h1>%s</h1>\n`, recipeName, recipeName)

	fmt.Fprintf(f, "<h4><a href = \"%s\">Original Recipe</a></h4>\n", url)

	if(len(recipe.PrepTime) != 0){
		fmt.Fprintf(f, "<h3>Prep Time: %s</h3>\n",recipe.PrepTime )
	}

	if(len(recipe.CookTime) != 0){
		fmt.Fprintf(f, "<h3>Cook Time: %s</h3>\n",recipe.CookTime )
	}

	if(len(recipe.TotalTime) != 0){
		fmt.Fprintf(f, "<h3>Total Time: %s</h3>\n",recipe.TotalTime )
	}


	f.WriteString("\n<h2>Ingredients</h2>\n")
	f.WriteString("<ul>\n")
	for _, ingredient := range recipe.Ingredients {
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

	for _, step := range recipe.Instructions {
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
	
	recipe := Recipe{}

	if graph, hasKey := jsonRecipe["@graph"]; hasKey {
		for _, item := range graph.([]any) {
			field := item.(map[string]any)
			recipe.GetIngredients(field)
			recipe.GetInstructions(field)
			recipe.GetCookTime(field)
			recipe.GetPrepTime(field)
			recipe.GetTotalTime(field)
		}
	} else{
		recipe.GetIngredients(jsonRecipe)
		recipe.GetInstructions(jsonRecipe)
		recipe.GetCookTime(jsonRecipe)
		recipe.GetPrepTime(jsonRecipe)
		recipe.GetTotalTime(jsonRecipe)
	}

	if recipe.Ingredients == nil {
		fmt.Fprintln(os.Stderr, "Failed to get Ingredients")
		os.Exit(1)

	}

	if recipe.Instructions == nil {
		fmt.Fprintln(os.Stderr, "Failed to get Instructions")
		os.Exit(1)
	}

	if generateHtml {
		err = ConstructHtml(url, recipeName, recipe)
	} else{
		err = ConstructMD(url, recipeName, recipe)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
