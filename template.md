# Шаблоны

## Визуализация шаблонов

`Context#Render(code int, name string, data interface{}) error` отображает шаблон с данными и отправляет текстовый / html\-ответ с кодом состояния. Шаблоны могут быть зарегистрированы путем настройки `Echo.Renderer` , что позволяет нам использовать любой шаблонизатор.

Пример ниже показывает, как использовать Go `html/template` :

1.  Реализовать `echo.Renderer` интерфейс

    ```go
    type Template struct {
        templates *template.Template
    }

    func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
        return t.templates.ExecuteTemplate(w, name, data)
    }

    ```

    копия

2.  Шаблоны предварительной компиляции

    `public/views/hello.html`

    ```html
    {{define "hello"}}Hello, {{.}}!{{end}}

    ```

    копия

    ```go
    t := &Template{
        templates: template.Must(template.ParseGlob("public/views/*.html")),
    }

    ```

    копия

3.  Зарегистрировать шаблоны

    ```go
    e := echo.New()
    e.Renderer = t
    e.GET("/hello", Hello)

    ```

    копия

4.  Визуализация шаблона внутри вашего обработчика

    ```go
    func Hello(c echo.Context) error {
        return c.Render(http.StatusOK, "hello", "World")
    }

    ```

    копия

### Advanced \- Calling Echo из шаблонов

В определенных ситуациях может быть полезно генерировать URI из шаблонов. Для этого вам нужно позвонить `Echo#Reverse` из самого шаблона. `html/template` Пакет Golang не совсем подходит для этой работы, но это можно сделать двумя способами: путем предоставления общего метода для всех объектов, передаваемых в шаблоны, или путем передачи `map[string]interface{}` и дополнения этого объекта в пользовательском рендерере. Учитывая гибкость последнего подхода, вот пример программы:

`template.html`

```html
<html>
    <body>
        <h1>Hello {{index . "name"}}</h1>

        <p>{{ with $x := index . "reverse" }}
           {{ call $x "foobar" }} &lt;-- this will call the $x with parameter "foobar"
           {{ end }}
        </p>
    </body>
</html>

```

копия

`server.go`

```go
package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
  e := echo.New()
  renderer := &TemplateRenderer{
      templates: template.Must(template.ParseGlob("*.html")),
  }
  e.Renderer = renderer

  // Named route "foobar"
  e.GET("/something", func(c echo.Context) error {
      return c.Render(http.StatusOK, "something.html", map[string]interface{}{
          "name": "Dolly!",
      })
  }).Name = "foobar"

  e.Logger.Fatal(e.Start(":8000"))
}
```
