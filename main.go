package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type PortfolioItem struct {
	ImagePath   string
	Name        string
	Description string
}

func main() {
	var config map[string]interface{}

	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("can't open config.yml: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("yaml unmarshal: %v", err)
	}

	config["year"] = time.Now().Year()
	style := "default"
	if val, found := config["style"]; found {
		style = val.(string)
	}

	tpl, err := template.ParseGlob("./tpl/" + style + "/*")
	if err != nil {
		log.Fatalf("can't scan template files: %v", err)
	}

	posts := createPosts()

	config["items"] = posts

	// create index
	err = createFileFromTemplate("index.html", "index.html", config, tpl)
	if err != nil {
		log.Fatalf("can't create file: %v", err)
	}

	// create posts
	for _, post := range posts {
		config["post"] = post
		err = createFileFromTemplate("post.html", fmt.Sprintf("./posts/%s/index.html", post.Path), config, tpl)
		if err != nil {
			log.Fatalf("can't create file: %v", err)
		}
	}

	fmt.Println("Build static complete!")

	if os.Getenv("run") == "true" {
		fmt.Println("Running web server on http://localhost:8080/")
		r := gin.Default()
		r.Static("/dist", "./dist")
		r.Static("/posts", "./posts")
		// Listen and serve on 0.0.0.0:8080
		r.GET("/", func(c *gin.Context) {
			c.File("./index.html")
		})
		r.Run(":8081")
	}
}

type Post struct {
	Path        string
	Title       string
	Description string
	BasePhoto   string `yaml:"base_photo"`
	Photos      []string
	Draft       bool
}

func createPosts() []Post {
	dirs, err := ioutil.ReadDir("./posts/")
	if err != nil {
		log.Fatal(err)
	}
	var posts []Post
	for _, f := range dirs {
		postPath := f.Name()
		post := Post{}
		postConfPath := fmt.Sprintf("./posts/%s/post.yml", postPath)
		yamlFile, err := ioutil.ReadFile(postConfPath)
		if err != nil {
			fmt.Println("not found post config in post ", postPath)
			continue
		}

		err = yaml.Unmarshal(yamlFile, &post)
		if err != nil {
			fmt.Println("can't unmarshall post ", postPath)
			continue
		}

		if post.Draft {
			// не рендерим черновик
			continue
		}
		post.Path = postPath
		posts = append(posts, post)
	}

	return posts
}

func createFileFromTemplate(templateName, fileName string, data map[string]interface{}, tpl *template.Template) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tpl.ExecuteTemplate(f, templateName, data)
	if err != nil {
		return err
	}
	return nil
}
