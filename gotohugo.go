//go:directive to be ignored by gotohugo
/*
+++
title = "gotohugo: Converting commented Go files to Markdown with custom Hugo shortcuts"
author = "Christoph Berger"
date = "2016-04-14"
categories = ["tool"]
tags = ["Go", "Hugo", "Markdown", "Hype"]
+++


`gotohugo` converts a .go file into a Markdown file. Comments can (and should) contain [Markdown](https://daringfireball.net/projects/markdown) text. Comment delimiters are stripped, and Go code is put into code fences.

<!--more-->

Extra #1: A non-standard "HYPE" tag can be used for inserting Tumult Hype HTML animations. This tag resembles an image tag but with the "!" replaced by "HYPE", like: `HYPE[Description](path/to/exported_hype.html)`. It is replaced by the corresponding HTML snippet that loads the animation. To create the anmiation files, export your Tumult Hype animation to HTML5 and ensure the "Also save HTML file" checkbox is checked. `gotohugo` then extracts the required HTML snippet from the file and copies the `hyperesources` directory to the output folder.

Extra #2: gotohugo inserts Hugo shortcodes around doc and code parts to help creating a side-by-side layout à la docgo.


## Usage

	gotohugo [-outdir "path/to/outputDir"] [-nocopy] <gofile.go>

### Flags

*`-outdir`: Specifies the output directory. Defaults to "out".
*`-nocopy`: If set, the image and animation files do not get copied to the output directory.
*`-subdir`: If set, the image and animation files are copied into &lt;outdir>/&lt;subdir>, rather than into &lt;outdir>.

## License

(c) 2016 Christoph Berger. All Rights Reserved.
This code is governed by a BSD 3-clause license that can be found in LICENSE.txt.

*/

// ## Imports and Globals

package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	commentPtrn      = `^\s*//\s?`
	commentStartPtrn = `^\s*/\*\s?`
	commentEndPtrn   = `\s?\*/\s*$`
	directivePtrn    = `^//go:`
	imagePtrn        = `[^\x60]!\[[^\]]+\]\( *([^"\)]+) *["\)]` // \x60 = backtick
	hypePtrn         = `[^\x60]HYPE\[[^\]]+\]\( *([^\)]+) *\)`
)

var (
	comment          = regexp.MustCompile(commentPtrn)      // pattern for single-line comments
	commentStart     = regexp.MustCompile(commentStartPtrn) // pattern for /* comment delimiter
	commentEnd       = regexp.MustCompile(commentEndPtrn)   // pattern for */ comment delimiter
	directive        = regexp.MustCompile(directivePtrn)    // pattern for //go: directive, like //go:generate
	imageTag         = regexp.MustCompile(imagePtrn)        // pattern for Markdown image tag
	hypeTag          = regexp.MustCompile(hypePtrn)         // pattern for Hype animation tag
	allCommentDelims = regexp.MustCompile(commentPtrn + "|" + commentStartPtrn + "|" + commentEndPtrn)
	outDir           = flag.String("outdir", "out", "Output directory")
	dontCopyMedia    = flag.Bool("nocopy", false, "Do not copy media files to outdir")
	subDir           = flag.Bool("subdir", false, "Use subdirectory <outdir>/<gofilebasename>/ for media files, ex.: out/gotohugo/")
)

// ## First, some helper functions
//
// copyFiles copies a list of files or directories to a destination directory.
// The destination path must exist.
// The source paths must be relative. (Usually they are, as they are taken from an MD image tag)
func copyFiles(dest string, srcpaths map[string]struct{}) (err error) {
	var result []byte
	for src, _ := range srcpaths {
		result, err = exec.Command("cp", "-R", path.Clean(strings.Trim(src, " \t")), path.Clean(path.Join(dest, src))).Output() // TODO: Windows "copy"
		if err != nil {
			return errors.New(string(result) + "\n" + err.Error())
		}
	}
	return nil
}

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

// extractMediaPath receives a line of text and searches for an image
// tag. If it finds one, it adds the path to the media list.
// NOTE: The function can only handle one image tag per line.
func extractMediaPath(line string) (path string, err error) {
	matches := imageTag.FindStringSubmatch(line)
	if len(matches) == 0 {
		return "", nil
	}
	if len(matches) == 1 {
		return "", errors.New("Error: Found image tag but no valid path, in line:\n" + line)
	}
	return strings.Trim(matches[1], " \t"), nil
}

// imageTag should properly match the following image tags:
//
// `![Alt text](gotohugo_animation.gif)`
//
// ![Alt text](gotohugo_animation.gif)
//
// `![Alt text](gotohugo_animation.gif "Title")` (With image title)
//
// ![Alt text](gotohugo_animation.gif "Title")
//
// `![Alt text](gotohugo image.jpg)` (With a space in the path)
//
// ![Alt text](gotohugo image.jpg)
//
// `![Alt text](gotohugo image.jpg "Title")`  (With space and title)
//
// ![Alt text](gotohugo image.jpg "Title")

// getHTMLSnippet opens the file determined by `path`, and scans the file for the HTML
// snippet to insert. It returns the HTML snippet.
func getHTMLSnippet(path string) (out string, err error) {
	hypeHTML, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.New("Unable to open Hype file " + path + "\n" + err.Error())
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
			out += strings.Trim(line, "	\t") + "\n"
		}
	}
	return out + "\n", nil
}

// replaceHypeTag identifies a tag like `HYPE[description](gotohugo_animation.html)`
// and replaces it by the correspoding HTML snippet generated by [Tumult Hype](http://tumult.com).
//
// HYPE[description](gotohugo_animation.html)
//
// It returns the (possibly modified) line and the path to the hyperesources directory.
func replaceHypeTag(line string) (out string, path string, err error) {
	matches := hypeTag.FindStringSubmatch(line)
	if len(matches) == 0 {
		return line, "", nil
	}
	if len(matches) == 1 {
		return "", "", errors.New("Error: Found Hype tag but no valid path, in line:\n" + line)
	}
	path = matches[1]
	out, err = getHTMLSnippet(path)
	out += "<noscript><em>Please enable JavaScript to view the animation.</em></noscript>\n"
	path = strings.Replace(path, ".html", ".hyperesources", -1)
	return out, path, err
}

// convert receives a string containing commented Go code and converts it
// line by line into a Markdown document. Collect and return any media files
// found during this process.
func convert(in string) (out string, media map[string]struct{}, err error) {
	const (
		neither = iota
		comment
		code
	)
	lastLine := neither
	media = map[string]struct{}{}

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
			// Detect `![image](path)` tags and add the path to the
			// media list.
			path, err := extractMediaPath(line)
			if err != nil {
				return "", nil, errors.New("Unable to extract media path from line " + line + "\n" + err.Error())
			}
			if path != "" {
				media[path] = struct{}{}
			}

			repl, path, err := replaceHypeTag(line)
			if err != nil {
				return "", nil, errors.New("Failed generating Hype tag from line " + line + "\n" + err.Error())
			}
			if repl != "" && path != "" {
				out += repl
				media[path] = struct{}{}
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
	return out, media, nil
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

// `createPath` receives a path and creates all directories in that path
// that are missing. Permissions are set to u+rwx go+r.
func createPath(p string) (err error) {
	if path.Clean(p) != "." {
		err = os.MkdirAll(p, 0744) // -rwxr--r--
		if err != nil {
			return errors.New("Cannot create path: " + p + " - Error: " + err.Error())
		}
	}
	return nil
}

// ### Now the actual conversion
//
// `convertFile` takes a file name, reads that file, converts it to
// Markdown, and writes it to `*outDir/&lt;basename>.md
func convertFile(filename string) (media map[string]struct{}, err error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Cannot read file " + filename + "\n" + err.Error())
	}
	name := filepath.Base(filename)
	ext := ".md"
	basename := base(name) // strip ".go"
	outname := filepath.Join(*outDir, basename) + ext
	md, media, err := convert(string(src))
	if err != nil {
		return nil, errors.New("Error converting " + filename + "\n" + err.Error())
	}
	err = createPath(*outDir)
	if err != nil {
		return nil, err // The error message from createPath is chatty enough.
	}
	err = ioutil.WriteFile(outname, []byte(md), 0644) // -rw-r--r--
	if err != nil {
		return nil, errors.New("Cannot write file " + outname + " \n" + err.Error())
	}
	return media, nil
}

// ## main - Where it all starts

func main() {
	flag.Parse()
	for _, filename := range flag.Args() {
		log.Println("Converting", filename)
		media, err := convertFile(filename)
		if err != nil {
			log.Fatal("[Conversion Error] " + err.Error())
		}
		if *dontCopyMedia == false && media != nil && (path.Clean(*outDir) != "." || *subDir == true) {
			log.Println("Copying media")
			out := *outDir
			if *subDir {
				out = filepath.Join(*outDir, base(filename))
				err := createPath(out)
				if err != nil {
					log.Fatal("[CopyMedia Error] Cannot create subdir for media files.\n" + err.Error())
				}
			}
			err := copyFiles(out, media)
			if err != nil {
				log.Fatal("[CopyMedia Error] cp failed:\n" + err.Error())
			}
		}
	}
	log.Println("Done.")
}
