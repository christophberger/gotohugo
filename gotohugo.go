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

3. Because of 1., gotohugo tries to find any Hype animation hmtl file in `outputDir/basename/hypename.html`. Gotohugo needs this file to extract the HTML snippet that replaces the HYPE tag. If gotohugo does not find the animation HTML that the HYPE tag points to, it subtitutes a warning message that will be visible on the rendered page.

## How to write proper gotohugo-friendly code documents

### Document sections and comment/code sections

Comments and code shall get rendered side-by-side if the screen width allows. Pure documentation, on the other hand, shall be rendered as a single column, centered to the screen and with optimal reading width (about 30em).

To distinguish between pure documentation and comment/code pairs without the need for extra markup, the following rules apply:

### Documents are `/``*` comment regions `*``/`.

Any "pure" document section, especially the very first one, **must** be enclosed in multiline comment delimiters.

### Comment/code pairs must use // for comments.

No multiline comment delimiters allowed here.
This way, gotohugo can easily detect the different section types and create the relevant output without ever having to go back to previous lines.
Also, the author does not need to memorize any kind of special markup syntax, nor insert any additional keywords into the document.

A line comment **must** be followed by code. Otherwise, use a multiline comment instead.

### Add Hugo front matter right at the beginning.

After an optional //go:... directive and the beginning of the first multiline comment delimiter, add the necessary Hugo front matter.

Front matter **must** exist. Hugo cannot process a post properly without front matter. `gotohugo` fails processing the source file if it contains no front matter.
Use the toml or yaml syntax, depending on the setting in the Hugo configuration.

### Add a summary divider.

The first part of the intro is a summary that Hugo can render on the list page. To mark the end of the summary, use the Hugo summary divider to manually define where the article gets split:

`<!``--more-->`

After that, continue with the intro.

The summary divider must exist exactly once in this document.

### Images are placed in a subfolder.

By convention, images and animation files are placed in a subfolder that has the basename of the markdown file.

For example, if the markdown file is named `gotohugo.md`, then the images and animations must be placed in the subfolder `gotohugo`. This subfolder is in the same folder as `gotohugo.go`.

### Images and Hype animations MUST exist at the output dir, in the aforementioned subfolder.

Reason is that `gotohugo` fetches an HTML snippet from the Hype HTML. If it cannot find the Hype HTML, it erros out.


### Do not specify the path of an image or animation html.

`gotohugo` automatically expands image and animation references as required.

Example:

`![image](image.png)` gets expanded to `![image](/post/gotohugo/image.png)`

### Example of a gotohugo-friendly source code file.

Examine `gotohugo.go`, which follows all the above rules and conventions.


## TODO

[] Replace strings with []byte where this can help avoiding excessive copying & garbage creating.


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
	preformatPtrn    = `\x60|^ {4,}|^\t\s*` // \x60 = backtick
	commentPtrn      = `^\s*//\s?`
	commentStartPtrn = `^\s*/\*\s?`
	commentEndPtrn   = `\s?\*/\s*$`
	frontmatterPtrn  = `^\s*(\+\+\+)|(---)\s*$`
	imagePtrn        = `(!\[[^\]]+\]\( *)([^"\)]*?)(.*?\))`
	hypePtrn         = `HYPE\[[^\]]+\]\( *([^\)]+) *\)`
	srcPtrn          = `(src=")(.*\.hyperesources/)`
)

var (
	preformat        = regexp.MustCompile(preformatPtrn)    // matches preformatted text
	commentRe        = regexp.MustCompile(commentPtrn)      // matches single-line comments
	commentStart     = regexp.MustCompile(commentStartPtrn) // matches /* comment delimiter
	commentEnd       = regexp.MustCompile(commentEndPtrn)   // matches */ comment delimiter
	frontmatterDelim = regexp.MustCompile(frontmatterPtrn)  // matches Hugo front matter delimiters
	imageTag         = regexp.MustCompile(imagePtrn)        // matches Markdown image tag
	hypeTag          = regexp.MustCompile(hypePtrn)         // matches Hype animation tag
	srcTag           = regexp.MustCompile(srcPtrn)          // matches Hype container div src tag
	allCommentDelims = regexp.MustCompile(commentPtrn + "|" + commentStartPtrn + "|" + commentEndPtrn)
	outDir           = flag.String("out", "out", "Output directory")
)

// ## First, some helper functions
//
// isLineComment returns true if the text in the input string starts with //.
func isLineComment(line string) bool {
	if commentRe.FindString(line) != "" {
		// "//" Comment line found.
		return true
	}
	return false
}

// isCommentStart detects the start of a multiline comment.
func isCommentStart(line string) bool {
	if commentStart.FindString(line) != "" {
		return true
	}
	return false
}

// isCommentEnd detects the end of a multiline comment.
func isCommentEnd(line string) bool {
	if commentEnd.FindString(line) != "" {
		return true
	}
	return false
}

// isFrontmatterDelim receives an integer and increases it by one
// if it finds a frontmatter deliminter in the current line.
func isFrontmatterDelim(line string) bool {
	if frontmatterDelim.FindString(line) != "" {
		return true
	}
	return false
}

// isSummaryDivider detects the summary divider.
func isSummaryDivider(line string) bool {
	if strings.Index(line, "<!--more-->") > -1 {
		return true
	}
	return false
}

func isPreformatted(line string) bool {
	return preformat.FindString(line) != ""
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
	if isPreformatted(line) {
		return line
	}
	return string(imageTag.ReplaceAll([]byte(line), []byte("$1"+extendPath("$2", basename)+"$3")))
}

/*
imageTag should properly match the following image tags:

`![Animation GIF](animation.gif)`

![Animation GIF]( animation.gif )

(Same but with spaces around the path:) ![Animation GIF with spaces]( animation.gif )

`![Animation GIF with title](animation.gif "Title")` (With image title)

![Animation GIF with title](animation.gif "Title")

    ![Image with space in path](an image.png) (With a space in the path)

![Image with space in path](an image.png)

	Same but with title: ![With space and title](an image.png "Title")

![With space and title](an image.png "Title")
*/

// getHTMLSnippet opens the file determined by `path`, and scans the file for the HTML
// snippet to insert. It returns the HTML snippet.
func getHTMLSnippet(path, basename string) (out string) {
	hypeHTML, err := ioutil.ReadFile(path)
	if err != nil {
		wrappedErr := errors.Wrap(err, "**No Hype file found at "+path+". Please run gohugo again after creating the Hype animation HTML export.")
		log.Println(wrappedErr.Error()) // notify the developer via shell
		return wrappedErr.Error()       // remind the developer by adding the message to the rendered page
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
//
// It returns:
// * out: the (possibly modified) line
// * found: true if a HYPE tag was found (and processed)
func replaceHypeTag(line, base string) (out string, found bool, err error) {
	// Do not process preformatted text
	if isPreformatted(line) {
		return line, false, nil
	}
	// Find the HYPE tag if it exists.
	matches := hypeTag.FindStringSubmatch(line)
	if len(matches) == 0 {
		return line, false, nil
	}
	if len(matches) < 2 {
		return "", false, errors.New("Error: Found Hype tag but no valid path, in line:\n" + line)
	}
	// substitute the Hype HTML snippet for the HYPE tag.
	path := matches[1]
	out = getHTMLSnippet(filepath.Join(*outDir, base, path), base)
	out += "<noscript class=\"nohype\"><em>Please enable JavaScript to view the animation.</em></noscript>\n"
	return out, true, err
}

/*
HYPE[description](animation.html)
*/

// div returns a Hugo shortcode of the form
// &#123;{% div <name> %}}.
func div(name string) string {
	return "{" + "{% div " + name + " %}}\n"
}

// divEnd returns the end marker of a div.
func divEnd(name string) string {
	return "{" + "{% divend %}} <!--" + name + "-->\n"
}

// convert receives a string containing commented Go code and converts it
// line by line into a Markdown document.
func convert(in, base string) (out string) {
	const (
		beforefrontmatter = iota
		frontmatter
		summary
		intro
		doc
		comment
		code
		none
	)
	status := beforefrontmatter

	// Turn CR/LF line endings into pure LF line endings.
	in = strings.Replace(in, "\r", "", -1)
	// Split at newline and process each line.
	for _, line := range strings.Split(in, "\n") {

		// First we do some line processing that does **not** necessarily call
		// `continue`.

		// Images and Hype animations can be located in the intro,
		// in comments, or in pure doc sections.
		if status == doc || status == comment || status == intro {

			// If the line contains an image tag, extend the path of the tag.
			line = extendImagePath(line, base)

			// If the line contains a Hype tag, replace it with the Hype HTML snippet.
			line, found, err := replaceHypeTag(line, base)
			if err != nil {
				e := errors.Wrap(err, "Failed generating Hype tag from line "+line)
				errors.Print(e)
				out += e.Error()
			}
			if found {
				out += line
				continue
			}
		}

		// if the line belongs to Hugo front matter, append it to out
		// and continue with the next line.
		if status == beforefrontmatter {
			if isFrontmatterDelim(line) { // start of front matter.
				status = frontmatter
				out += line + "\n"
				continue
			}
			// Discard anything before the front matter. There should **only**
			// be an optional //go:... directive, and the start of the first
			// multiline comment, and nothing else.
			continue
		}

		// Within frontmatter, if the second delimiter is found,
		// switch to summary section.
		// Also generate a `gotohugo` namespace div.
		if status == frontmatter {
			out += line + "\n"
			if isFrontmatterDelim(line) { // end of front matter. Summary section begins.
				status = summary
				out += div("gotohugo")
				out += div("summary doc")
				continue
			}
		}

		// After the summary divider, start the intro.
		if status == summary {
			if isSummaryDivider(line) {
				out += divEnd("summary doc")
				out += line + "\n"
				out += div("intro doc")
				status = intro
				continue
			}
			out += line + "\n"
			continue
		}

		// Intro is finished when the comment end delimiter occurs.
		// The status afterwards is not defined. Comment/code pairs might follow,
		// or another multiline comment. Or the end of the file.
		if status == intro {
			if isCommentEnd(line) {
				out += divEnd("intro doc")
				status = none
				continue
			}
			out += line + "\n"
			continue
		}

		// A line comment can occur after code, after another line comment,
		// or when no other section is active.
		if status == none || status == code {
			if isLineComment(line) {
				// If the last line was code, add a closing code fence.
				if status == code {
					out += "```\n\n"
					out += divEnd("code")
					out += divEnd("ccpair")
					out += div("ccpair")
				}
				// Multiline comments switch the status to none at the end.
				// In this case, start a new source section.
				if status == none {
					out += div("source")
					out += div("ccpair")
				}
				status = comment
				out += div("comment")
				// Strip the comment delimiters.
				out += commentRe.ReplaceAllString(line, "") + "\n"
				continue
			}
		}

		// While processing line comments.
		if status == comment {
			// If still looking at a line comment, strip the delims.
			// Else switch into code status.
			if isLineComment(line) {
				out += commentRe.ReplaceAllString(line, "") + "\n"
				continue
			} else {
				status = code
				out += divEnd("comment")
				out += div("code")
				out += "\n```go\n"
				out += line + "\n"
				continue
			}
		}

		// While processing code, look out for comments.
		if status == code {

			// A line comment occurs. End the code section.
			if isLineComment(line) {
				status = comment
				out += "```\n\n"
				out += divEnd("code")
				out += divEnd("ccpair")
				out += div("ccpair")
				out += div("comment")
				out += commentRe.ReplaceAllString(line, "") + "\n"
				continue
			}

			// A multiline commment starts. End the code section and switch to
			// single-column layout by closing the "source" div.
			if isCommentStart(line) {
				status = doc
				out += "```\n\n"
				out += divEnd("code")
				out += divEnd("ccpair")
				out += divEnd("source")
				out += div("doc")
				out += commentStart.ReplaceAllString(line, "") + "\n"
				continue
			}
			out += line + "\n"
			continue

		}

		// At the end of a multline comment, we don't know for sure
		// what comes next, so we set the status to none.
		if status == doc {
			if isCommentEnd(line) {
				out += divEnd("doc")
				status = none
				continue
			}
			out += line + "\n"
			continue
		}

		// Outside any status? Just pass the line to the output.
		if status == none {
			out += line + "\n"
		}
	}

	// The last line in the file might be code.
	// We need a closing code fence then, and we need to close the divs, too.
	if status == code {
		out += "\n```\n"
		out += divEnd("code")
		out += divEnd("ccpair")
	}

	// Close the `gotohugo` namespace div.
	out += divEnd("gotohugo")

	return out
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
	md := convert(string(src), basename)
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
