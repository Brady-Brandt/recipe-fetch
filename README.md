# recipe-fetch
A simple CLI tool that extracts recipes from a webpage and converts them into a Markdown file containing ingredients and instructions.

## Features

- Extract recipe data from supported recipe websites*
- Generate clean Markdown output
- Save recipes for offline use
- Simple command-line interface

> [!NOTE]  
> *Right now this only includes sites that store the recipes in json. Even those are not guarenteed to work

## Installation

### Build from source

```bash
git clone https://github.com/Brady-Brandt/recipe-fetch.git
cd recipe-fetch
go build
```

### Run directly
This will try to pull the recipe from the url and if successful it will create a markdown file `recipe_name.md`
```bash
go run . -url https://example.com/recipe -o recipe_name
```

## Usage
```bash
recipe-fetch -url <recipe-url> [-o <recipe-name>]
```
If no recipe name is inputted, the last part of the url will be used instead.
The following will create a file `homemade-brownies.md`
```bash
recipe-fetch -url "https://example.com/homemade-brownies
```
## Example Output
```markdown
# homemade-brownies

## Ingredients

- 2 cups flour
- 1 cup sugar
- 2 eggs

## Instructions
### Step 1
Preheat oven to 350°F (175°C).
### Step 2
Mix ingredients.
### Step 3
Bake for 12 minutes.
```
