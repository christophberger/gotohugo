//go:directive to be ignored by gotohugo
/*
+++
title = "gotohugo: Converting commented Go files to Markdown with custom Hugo shortcuts"
author = "Christoph Berger"
date = "2016-04-25"
categories = ["Blogging"]
tags = ["Go", "Hugo", "Markdown", "Hype"]
+++


`gotohugo` converts a .go file into a Markdown file. Comments can (and should) contain [Markdown](https://daringfireball.net/projects/markdown) text. Comment delimiters are stripped, and Go code is put into code fences. There are also two extra features included for free.

<!--more-->

Extra #1: A non-standard "HYPE" tag can be used for inserting Tumult Hype HTML animations. This tag resembles an image tag but with the "!" replaced by "HYPE", like: `HYPE[Description](path/to/exported_hype.html)`. It is replaced by the corresponding HTML snippet that loads the animation. To create the anmiation files, export your Tumult Hype animation to HTML5 and ensure the "Also save HTML file" checkbox is checked. `gotohugo` then extracts the required HTML snippet from the file and copies the `hyperesources` directory to the output folder.

Extra #2: gotohugo inserts Hugo shortcodes around doc and code parts to help creating a side-by-side layout à la docgo, where the code comments appear in an extra column left to the code. This very much adds to readability IMHO. This feature also comes with full Responsive Layout capability - if the viewport is too narrow, code and comment collapse into a single column.


## Usage

	gotohugo [-o "path/to/outputDir"] <gofile.go>

### Flags

*`-o`: Specifies the output directory. Defaults to `./out`. The path must already exist. By convention it is the path to Hugo's `content/post/` directory.

## Notes

1. Unlike gotomarkdown, gotohugo does not handle any media files itself. All media files must be available at the output destination, in a subdirectory whose name is the base name of the go file.
   Example: mytutorial.go gets turned into /post/mytutorial.md, and all meda files then must reside in /post/mytutorial/...
   The point here is that right now, Hugo does not create subdirectories for posts; they all are created in `<hugo>/content/post`. To reduce clutter, all media files related to a post should therefore be put into a subdirectory of the post's base name.
   As far as Hugo is concerned, this is just a convention; however, gotohugo relies on this file structure.

2. To play nice with the Permalink feature of Hugo, gotohugo automatically creates the full path to the image file, starting from the content directory. That is, if your image resides in `content/post/mypost/myimage.jpg`, and your Markdown tag is like, `[My Image](myimage.jpg)`, gotohugo expands the tag to `[My Image](/post/mypost/myimage.jpg`.

3. Because of 1., gotohugo tries to find any Hype animation hmtl file in `outputDir/basename/hypename.html`. Gotohugo needs this file to extract the HTML snippet that replaces the HYPE tag. If gotohugo does not find the animation HTML that the HYPE tag points to,

## License

(c) 2016 Christoph Berger. All Rights Reserved.
This code is governed by a BSD 3-clause license that can be found in LICENSE.txt.

*/

// ## Imports and Globals

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	commentPtrn      = `^\s*//\s?`
	commentStartPtrn = `^\s*/\*\s?`
	commentEndPtrn   = `\s?\*/\s*$`
	directivePtrn    = `^//go:`
	imagePtrn        = `([^\x60]!\[[^\]]+\]\( *)([^\)]+\))` // \x60 = backtick
	hypePtrn         = `[^\x60]HYPE\[[^\]]+\]\( *([^\)]+) *\)`
	srcPtrn          = `(src=")(.*\.hyperesources/)`
)

var (
	comment          = regexp.MustCompile(commentPtrn)      // pattern for single-line comments
	commentStart     = regexp.MustCompile(commentStartPtrn) // pattern for /* comment delimiter
	commentEnd       = regexp.MustCompile(commentEndPtrn)   // pattern for */ comment delimiter
	directive        = regexp.MustCompile(directivePtrn)    // pattern for //go: directive, like //go:generate
	imageTag         = regexp.MustCompile(imagePtrn)        // pattern for Markdown image tag
	hypeTag          = regexp.MustCompile(hypePtrn)         // pattern for Hype animation tag
	srcTag           = regexp.MustCompile(srcPtrn)          // pattern for Hype container div src tag
	allCommentDelims = regexp.MustCompile(commentPtrn + "|" + commentStartPtrn + "|" + commentEndPtrn)
	outDir           = flag.String("out", "out", "Output directory")
)

// ## First, some helper functions
//
// commentFinder returns a function that determines if the current line belongs to
// a comment region.
func commentFinder() func(string) bool {
	commentSectionInProgress := false
	return func(line string) bool {
		if comment.FindString(line) != "" {
			// "//" Comment line found.
			return true
		}
		// If the current line is at the start `/*` of a multi-line comment,
		// set a flag to remember we're within a multi-line comment.
		if commentStart.FindString(line) != "" {
			commentSectionInProgress = true
			return true
		}
		// At the end `*/` of a multi-line comment, clear the flag.
		if commentEnd.FindString(line) != "" {
			commentSectionInProgress = false
			return true
		}
		// The current line is within a `/*...*/` section.
		if commentSectionInProgress {
			return true
		}
		// Anything else is not a comment region.
		return false
	}
}

// isInComment returns true if the current line belongs to a comment region.
// A comment region `//` is either a comment line (starting with `//`) or
// a `/*...*/` multi-line comment.
var isInComment func(string) bool = commentFinder()

// isDirective returns true if the input argument is a Go directive,
// like `//go:generate`.
func isDirective(line string) bool {
	if directive.FindString(line) != "" {
		return true
	}
	return false
}

// extendPath takes a string that should contain a filename
// and prepends `/post/<basename>/` to it.
func extendPath(filename, basename string) string {
	return filepath.Join("/post", basename, filename)
}

// func extendSrc takes a string that should contain the line from the HTML snippet that
// starts with `<div id="animation_hype_container"...` and prepends `/post/<basename>` to
// the src="..." string.
func extendSrc(src, basename string) string {
	return string(srcTag.ReplaceAll([]byte(src), []byte("$1"+extendPath("$2", basename))))
}

// extendImagePath receives a line of text and searches for an image
// tag. If it finds one, it extends the image path to include
// `/post/<basename>/` and returns the modified line.
// Otherwise it returns the unmodified line.
func extendImagePath(line, basename string) string {
	return string(imageTag.ReplaceAll([]byte(line), []byte("$1"+extendPath("$2", basename))))
}

// imageTag should properly match the following image tags:
//
// `![Alt text](animation.gif)`
//
// ![Alt text](animation.gif)
// (Same but with spaces around the path:) ![Alt text]( animation.gif )
//
// `![Alt text](animation.gif "Title")` (With image title)
//
// ![Alt text](animation.gif "Title")
//
// `![Alt text](an image.png)` (With a space in the path)
//
// ![Alt text](an image.png)
//
// `![Alt text](an image.png "Title")`  (With space and title)
//
// ![Alt text](an image.png "Title")

// getHTMLSnippet opens the file determined by `path`, and scans the file for the HTML
// snippet to insert. It returns the HTML snippet.
func getHTMLSnippet(path, basename string) (out string) {
	hypeHTML, err := ioutil.ReadFile(path)
	if err != nil {
		return "**No Hype file found at " + path + "\nPlease run gohugo again after creating the Hype animation HTML export."
	}
	inSnippet := false
	// Remove carriage returns.
	lines := strings.Replace(string(hypeHTML), "\r", "", -1)
	// Split at newline and process each line.
	for _, line := range strings.Split(lines, "\n") {
		if strings.Index(line, "<!-- copy these lines to your document: -->") >= 0 {
			inSnippet = true
			continue
		}
		if strings.Index(line, "<!-- end copy -->") >= 0 {
			if inSnippet == true {
				break
			}
			inSnippet = false // there can be more than one "end copy" strings in the file
		}
		if inSnippet {
			out += extendSrc(strings.Trim(line, "	\t"), basename) + "\n"
		}
	}
	return out + "\n"
}

// replaceHypeTag identifies a tag like `HYPE[description](animation.html)`
// and replaces it by the correspoding HTML snippet generated by [Tumult Hype](http://tumult.com)
// through the "Export as HTML5 > Also save .html file" option.
//
// HYPE[description](animation.html)
//
// It returns the (possibly modified) line and the path to the hyperesources directory.
func replaceHypeTag(line, base string) (out string, path string, err error) {
	matches := hypeTag.FindStringSubmatch(line)
	if len(matches) == 0 {
		return line, "", nil
	}
	if len(matches) == 1 {
		return "", "", errors.New("Error: Found Hype tag but no valid path, in line:\n" + line)
	}
	path = matches[1]
	out = getHTMLSnippet(filepath.Join(*outDir, base, path), base)
	out += "<noscript class=\"nohype\"><em>Please enable JavaScript to view the animation.</em></noscript>\n"
	path = strings.Replace(path, ".html", ".hyperesources", -1)
	return out, path, err
}

// convert receives a string containing commented Go code and converts it
// line by line into a Markdown document.
func convert(in, base string) (out string, err error) {
	const (
		neither = iota
		comment
		code
	)
	lastLine := neither

	// Remove carriage returns.
	in = strings.Replace(in, "\r", "", -1)
	// Split at newline and process each line.
	for _, line := range strings.Split(in, "\n") {
		// Skip the line if it is a Go directive like //go:generate
		if isDirective(line) {
			continue
		}
		// Determine if the line belongs to a comment.
		if isInComment(line) {
			// Close the code block if a new comment begins.
			if lastLine == code {
				out += "```\n\n"
			}
			lastLine = comment

			// If the line contains an image tag, extend the path of the tag.
			line = extendImagePath(line, base)

			// If the line contains a Hype tag, replace it with the Hype HTML snippet.
			repl, path, err := replaceHypeTag(line, base)
			if err != nil {
				return "", errors.Wrap(err, "Failed generating Hype tag from line "+line)
			}
			if repl != "" && path != "" {
				out += repl
			} else {
				// Strip out any comment delimiter and add the line to the output.
				out += allCommentDelims.ReplaceAllString(line, "") + "\n"
			}
		} else { // not in comment
			// Open a new code block if the last line was a comment,
			// but take care of empty lines between two comment lines.
			if lastLine == comment && len(line) > 0 {
				lastLine = code
				out += "\n```go\n"
			}
			// Add code lines verbatim to the output.
			out += line + "\n"
		}
	}
	if lastLine == code {
		out += "\n```\n"
	}
	return out, nil
}

// ## Converting a file
//
// ### Again, some helper functions
//
// `base` strips the extension from a filename. For some reason, this
// function is missing from the standard path library.
func base(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// ### Now the actual conversion
//
// `convertFile` takes a file name, reads that file, converts it to
// Markdown, and writes it to `*outDir/<basename>.md`
// The path must already exist.
func convertFile(filename string) (err error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Cannot read file " + filename + "\n" + err.Error())
	}
	name := filepath.Base(filename)
	ext := ".md"
	basename := base(name) // strip ".go"
	outname := filepath.Join(*outDir, basename) + ext
	md, err := convert(string(src), basename)
	if err != nil {
		return errors.Wrap(err, "Error converting "+filename)
	}
	err = ioutil.WriteFile(outname, []byte(md), 0644) // -rw-r--r--
	if err != nil {
		return errors.Wrap(err, "Cannot write file "+outname)
	}
	return nil
}

// ## main - Where it all starts

func main() {
	flag.Parse()
	for _, filename := range flag.Args() {
		log.Println("Converting", filename)
		err := convertFile(filename)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Conversion Error"))
		}
	}
	log.Println("Done.")
}
