package templ

import (
	"context"
	"html"
	"io"

	"github.com/a-h/templ"
)

// Layout provides the base HTML layout for all pages
func Layout(title string, content templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>`+html.EscapeString(title)+`</title>
    <link rel="manifest" href="/static/pwa/manifest.json">
    <link rel="stylesheet" href="/static/css/main.css">
</head>
<body>
    <main>
`)
		if err != nil {
			return err
		}
		if err := content.Render(ctx, w); err != nil {
			return err
		}
		_, err = io.WriteString(w, `
    </main>
    <script src="/static/js/datastar.js"></script>
</body>
</html>`)
		return err
	})
}
