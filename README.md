# Path

Manage named paths (aka routes) used within your web application. This is NOT a router!

[![go report card](https://goreportcard.com/badge/github.com/joncalhoun/path "go report card")](https://goreportcard.com/report/github.com/joncalhoun/path)
[![Build Status](https://travis-ci.org/joncalhoun/path.svg?branch=master)](https://travis-ci.org/joncalhoun/path)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/joncalhoun/path?status.svg)](https://godoc.org/github.com/joncalhoun/path)

## Overview

The goal of the `path` package is to make it easier to manage named paths in a Go web application without having to rely on hard-coded paths. Here is a basic example:

```go
var pb path.Builder
pb.Set("edit_widget", "/widgets/:id/edit")

// ... in an http handler
http.Redirect(w, r, pb.Path("edit_widget", path.Params{
  "id": 123,
}), http.StatusFound) // redirects to "/widgets/123/edit"
```

You can also leverage the `path` package in your templates using the `FuncMap` method:

```go
var pb path.Builder
pb.Set("edit_widget", "/widgets/:id/edit")
tpl := template.New("").Funcs(pb.FuncMap()).Parse(`
  {{edit_widget_path (params "id" .Widget.ID)}}`)
data := struct {
  Widget Widget
}{
  Widget: Widget{
    ID: 123,
  },
}
tpl.Execute(w, data)
// Writes the following output:
//   /widgets/123/edit
```

## Installation

To install this package, simply `go get` it:

```
go get github.com/joncalhoun/form
```


## Use cases

This is much easier to understand with an example, so let's just jump right into one.

Imagine that I had a handler to process the creation of widgets. Once a widget is created I might want to redirect the user to a page where they can edit the widget's details. That redirect might look something like this before using the `path` package:

```go
func CreateWidget(w http.ResponseWriter, r *http.Request) {
  var widget Widget
  // ... create the widget

  path := fmt.Sprintf("/widgets/%v/edit", widget.ID)
  http.Redirect(w, r, path, http.StatusFound)
}
```

This works well enough, but relies on paths being hard coded into your application or defined as constants somewhere, and I often found myself writing the exact same code in every web application to turn paths in the format

You could try to fix this with constants, or by using the named routes in something like `gorilla/mux`, but I found that in nearly all web applications I tend to write the exact same path handling logic and typically didn't want to be too reliant on a specific router. The `path` package is intended to provide me with a single solution that can be easily used regardless of what router or whatever else I am using.

### Using path with other routers

It is also possible to use this along with another router, like [gorilla/mux](https://github.com/gorilla/mux), simply by replacing params with the format your router expects:

```go
var pb path.Builder
pb.Set("edit_widget", "/widgets/:id/edit")

r := mux.NewRouter()
r.HandleFunc(pb.Path("edit_widget", path.Params{
  "id": "{id:[0-9]+}",
}), EditWidgetHandler) // this generates the path /widgets/{id:[0-9]+}/edit
```

This works because the `path.Builder` simply replaces the ID with the value `{id:0-9]+}` which is a regex specifying the ID for `gorilla/mux`.


### Unset params are returned unchanged

If you define a URL param and don't provide a value for it the `path.Builder` will simply return the param as-is.

```go
var pb path.Builder
pb.Set("edit_widget", "/widgets/:id/edit")

pb.Path("edit_widget", nil) // returns /widgets/:id/edit
```
