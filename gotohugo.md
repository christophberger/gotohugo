+++
title = "gotohugo: Converting commented Go files to Markdown with custom Hugo shortcuts"
description = "gotohugo is a converter from .go to .md with some Hugo-specific additions. Comments are converted to Markdown text, code is converted to Markdown code blocks. Additional Hugo shortcodes are inserted for better layout control."
author = "Christoph Berger"
date = "2016-04-25"
domain = ["Blogging"]
categories = ["Tutorial"]
tags = ["Hugo", "Markdown", "Hype"]
+++
{{% div gotohugo %}}
{{% div summary doc %}}


`gotohugo` converts a .go file into a Markdown file. Comments can (and should) contain [Markdown](https://daringfireball.net/projects/markdown) text. Comment delimiters are stripped, and Go code is put into code fences. There are also two extra features included for free.

{{% divend %}} <!--summary doc-->
<!--more-->
{{< announcement >}}
{{% div intro doc %}}

Extra #1: A non-standard "HYPE" tag can be used for inserting Tumult Hype HTML animations. This tag resembles an image tag but with the "!" replaced by "HYPE", like: `HYPE[Description](path/to/exported_hype.html)`. It is replaced by the corresponding HTML snippet that loads the animation. To create the animation files, export your Tumult Hype animation to HTML5 and ensure the "Also save HTML file" checkbox is checked. `gotohugo` then extracts the required HTML snippet from the file and copies the `hyperesources` directory to the output folder.

Extra #2: gotohugo inserts Hugo shortcodes around doc and code parts to help creating a side-by-side layout à la docgo, where the code comments appear in an extra column left to the code. This very much adds to readability IMHO. This feature also comes with full Responsive Layout capability - if the viewport is too narrow, code and comment collapse into a single column.


## Usage

	gotohugo [-out "path/to/outputDir"] <gofile.go>
	gotohugo [-hugo "path/to/hugoRootDir"] <gofile.go>
	gotohugo [-watch "dir/to/watch"] [-out "path/to/outputDir"] [-v]
	gotohugo [-watch "dir/to/watch"] [-hugo "path/to/hugoRootDir"] [-v]

### Flags

*`-out`: Specifies the output directory. Defaults to `./out`. The path must already exist. By convention it is the path to Hugo's `content/post/` directory.
*`-hugo`: Specifies the Hugo root dir. Mutual exclusive to `-out`. When using `-hugo`, the output directory must point to the Hugo root directory. The markdown file will then be written to `<hugoRootDir>/content/post/<gofile.md>`. Hype files must already exist at `<hugoRootDir>/static/media/<gofile>/<hypefile>.html`, or else gotohugo fails replacing the HYPE tag with the corresponding Hype HTML.
*`-watch`: Watches the given directory. (Default: Current dir.) This must be the parent directory of one or more project directories. Gotohugo will only watch for changes to files whose names are the same as their directory, e.g., `gotohugo/gotohugo.go`. This is because each Hugo post is made from exactly one .go file, and this .go file must be named after its directory, to
distinguish it from other .go files that might also reside in the same dir but are not part of the blog post.
*`-d`: Debug-level logging.

### Precedence rules for flags and environment variables

* If either `-hugo` is used, or if `$HUGODIR` is set, `-out` has no effect.
* If neither of the flags nor `$HUGODIR` are set, output defaults to `./out/`.


## Notes

1. Unlike gotomarkdown, gotohugo does not handle any media files itself. All media files must be available at the output destination, in a subdirectory whose name is the base name of the go file.
   Example: mytutorial.go gets turned into content/post/mytutorial.md, and all media files then must reside in static/media/mytutorial/.
   The point here is that right now, Hugo does not create subdirectories for posts; they all are created in `<hugo>/content/post`. To reduce clutter, all media files related to a post should therefore be put into a subdirectory of the post's base name. Further, to avoid that Hugo grabs Hype HTML files and adds them to the list of posts, this subdirectory must reside outside the /post/ directory.
   As far as Hugo is concerned, this is just a convention; however, gotohugo relies on this file structure.

2. To play nice with the Permalink feature of Hugo, gotohugo automatically creates the full path to the image file, starting from the content directory. That is, if your image resides in `static/media/mypost/myimage.jpg`, and your Markdown tag is like, `[My Image](myimage.jpg)`, gotohugo expands the tag to `[My Image](/media/mypost/myimage.jpg`.

3. Because of 1., gotohugo tries to find any Hype animation hmtl file in `static/media/mypost/hypename.html`. Gotohugo needs this file to extract the HTML snippet that replaces the HYPE tag. If gotohugo does not find the animation HTML that the HYPE tag points to, it substitutes a warning message that will be visible on the rendered page.


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

**Note:** Anything before the front matter is **not** turned into Markdown. Put things like License remarks and other internal notes there.


### Add a summary divider.

The first part of the intro is a summary that Hugo can render on the list page. To mark the end of the summary, use the Hugo summary divider to manually define where the article gets split:

`<!``--more-->`

After that, continue with the intro.

The summary divider must exist exactly once in this document.


### Images are placed in a subfolder.

By convention, images and animation files are placed in a subfolder that has the base name of the markdown file.

For example, if the markdown file is named `gotohugo.md`, then the images and animations must be placed in the subfolder `gotohugo`. This subfolder is in the same folder as `gotohugo.go`.


### Images and Hype animations MUST exist at the output dir, in the aforementioned subfolder.

Reason is that `gotohugo` fetches an HTML snippet from the Hype HTML. If it cannot find the Hype HTML, it errors out.


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

{{% divend %}} <!--intro doc-->

{{% div source %}}
{{% div ccpair %}}
{{% div comment %}}
## Imports and Globals
{{% divend %}} <!--comment-->
{{% div code %}}

```go

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/pkg/errors"

	"github.com/google/gops/agent"
)

const (
	preformatPtrn    = `\x60|^ {4,}|^\t\s*` // \x60 = backtick
	commentPtrn      = `^\s*//\s?`
	commentStartPtrn = `^\s*/\*\s?`
	commentEndPtrn   = `\s*\*/\s*$`
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
	debug            = flag.Bool("d", false, "Enable debug-level logging.")
	watch            = flag.String("watch", "", "Watch dirs recursively. If <name>/<name>.go changes, convert the file to Hugo Markdown.")
	outDir           = flag.String("out", "out", "Output directory. Defaults to './out/'. If -hugo or $HUGODIR is set, -out has no effect.")
	hugoDir          = flag.String("hugo", "", "Hugo root directory. Overrides -out and $HUGODIR.")
	postDir          = "" // gets set to "/content/post" if -hugo is used instead of -out
	mediaDir         = "" // gets set to "/static/media" if -hugo is used instead of -out
	publicMediaDir   = "" // the media dir as the Web server sees it. Gets set to "/media" if -hugo is used.
)

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
## First, some helper functions

debug prints to the log output if the debug flag is set.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func dbg(args ...interface{}) {
	if *debug {
		log.Println(args)
	}
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
isLineComment returns true if the text in the input string starts with //.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func isLineComment(line string) bool {
	return commentRe.FindString(line) != ""
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
isCommentStart detects the start of a multiline comment.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func isCommentStart(line string) bool {
	return commentStart.FindString(line) != ""
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
isCommentEnd detects the end of a multiline comment.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func isCommentEnd(line string) bool {
	return commentEnd.FindString(line) != ""
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
isFrontmatterDelim receives an integer and increases it by one
if it finds a frontmatter deliminter in the current line.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func isFrontmatterDelim(line string) bool {
	return frontmatterDelim.FindString(line) != ""
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
isSummaryDivider detects the summary divider.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func isSummaryDivider(line string) bool {
	return strings.Contains(line, "<!--more-->")
}

func isPreformatted(line string) bool {
	return preformat.FindString(line) != ""
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
extendPath takes a string that should contain a filename
and prepends `/media/<basename>/` to it.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func extendPath(filename, basename string) string {
	return string(os.PathSeparator) + filepath.Join(publicMediaDir, basename, filename)
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
func extendSrc takes a string that should contain the line from the HTML snippet that
starts with `<div id="animation_hype_container"...` and prepends `/media/<basename>` to
the src="..." string.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func extendSrc(src, basename string) string {
	return string(srcTag.ReplaceAll([]byte(src), []byte("$1"+extendPath("$2", basename))))
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
extendImagePath receives a line of text and searches for an image
tag. If it finds one, it extends the image path to include
`/media/<basename>/` and returns the modified line.
Otherwise it returns the unmodified line.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func extendImagePath(line, basename string) string {
	if isPreformatted(line) {
		return line
	}
	return string(imageTag.ReplaceAll([]byte(line), []byte("$1"+extendPath("$2", basename)+"$3")))
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% divend %}} <!--source-->
{{% div doc %}}

imageTag should properly match the following image tags:

`![Animation GIF](animation.gif)`

![Animation GIF]( /gotohugo/animation.gif )

(Same but with spaces around the path:) ![Animation GIF with spaces]( /gotohugo/animation.gif )

`![Animation GIF with title](animation.gif "Title")` (With image title)

![Animation GIF with title](/gotohugo/animation.gif "Title")

    ![Image with space in path](an image.png) (With a space in the path)

![Image with space in path](/gotohugo/an image.png)

	Same but with title: ![With space and title](an image.png "Title")

![With space and title](/gotohugo/an image.png "Title")
{{% divend %}} <!--doc-->

{{% div source %}}
{{% div ccpair %}}
{{% div comment %}}
getHTMLSnippet opens the file determined by `path`, and scans the file for the HTML
snippet to insert. It returns the HTML snippet.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func getHTMLSnippet(path, basename string) (out string) {
	hypeHTML, err := ioutil.ReadFile(path)
	if err != nil {
		wrappedErr := errors.Wrap(err, "**No Hype file found at "+path+". Please run gotohugo again after creating the Hype animation HTML export.")
		log.Println(wrappedErr.Error()) // notify the developer via shell
		return wrappedErr.Error()       // remind the developer by adding the message to the rendered page
	}
	inSnippet := false
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Remove carriage returns.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	lines := strings.Replace(string(hypeHTML), "\r", "", -1)
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Split at newline and process each line.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	for _, line := range strings.Split(lines, "\n") {
		if strings.Contains(line, "<!-- copy these lines to your document: -->") {
			inSnippet = true
			continue
		}
		if strings.Contains(line, "<!-- end copy -->") {
			if inSnippet {
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

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
replaceHypeTag identifies a tag like `HYPE[description](animation.html)`
and replaces it by the corresponding HTML snippet generated by [Tumult Hype](http://tumult.com)
through the "Export as HTML5 > Also save .html file" option.


It returns:
* out: the (possibly modified) line
* found: true if a HYPE tag was found (and processed)
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func replaceHypeTag(line, base string) (out string, found bool, err error) {
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Do not process preformatted text
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	if isPreformatted(line) {
		return line, false, nil
	}
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Find the HYPE tag if it exists.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	matches := hypeTag.FindStringSubmatch(line)
	if len(matches) == 0 {
		return line, false, nil
	}
	if len(matches) < 2 {
		return "", false, errors.New("Error: Found Hype tag but no valid path, in line:\n" + line)
	}
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
substitute the Hype HTML snippet for the HYPE tag.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	path := matches[1]
	out = getHTMLSnippet(filepath.Join(*outDir, mediaDir, base, path), base)
	out += "<noscript class=\"nohype\"><em>Please enable JavaScript to view the animation.</em></noscript>\n"
	return out, true, err
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
div returns a Hugo shortcode of the form
&#123;{% div <name> %}}.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func div(name string) string {
	return "{" + "{% div " + name + " %}}\n"
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
divEnd returns the end marker of a div.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func divEnd(name string) string {
	return "{" + "{% divend %}} <!--" + name + "-->\n"
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
convert receives a string containing commented Go code and converts it
line by line into a Markdown document.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
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

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Turn CR/LF line endings into pure LF line endings.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	in = strings.Replace(in, "\r", "", -1)
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Split at newline and process each line.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	for _, line := range strings.Split(in, "\n") {

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
First we do some line processing that does **not** necessarily call
`continue`.
{{% divend %}} <!--comment-->
{{% div code %}}

```go

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Images and Hype animations can be located in the intro,
in comments, or in pure doc sections.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == doc || status == comment || status == intro {

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
If the line contains an image tag, extend the path of the tag.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
			line = extendImagePath(line, base)

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
If the line contains a Hype tag, replace it with the Hype HTML snippet.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
			line, found, err := replaceHypeTag(line, base)
			if err != nil {
				e := errors.Wrap(err, "Failed generating Hype tag from line "+line)
				fmt.Printf("%s\n", e)
				out += e.Error()
			}
			if found {
				out += line
				continue
			}
		}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
if the line belongs to Hugo front matter, append it to out
and continue with the next line.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == beforefrontmatter {
			if isFrontmatterDelim(line) { // start of front matter.
				status = frontmatter
				out += line + "\n"
				continue
			}
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Discard anything before the front matter. There should **only**
be an optional //go:... directive, and the start of the first
multiline comment, and nothing else.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
			continue
		}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Within front matter, if the second delimiter is found,
switch to summary section.
Also generate a `gotohugo` namespace div.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == frontmatter {
			out += line + "\n"
			if isFrontmatterDelim(line) { // end of front matter. Summary section begins.
				status = summary
				out += div("gotohugo")
				out += div("summary doc")
				continue
			}
		}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
After the summary divider, -
- insert the announcement shortcode
- start the intro.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == summary {
			if isSummaryDivider(line) {
				out += divEnd("summary doc")
				out += line + "\n"
				out += "{" + "{< announcement >}}\n"
				out += div("intro doc")
				status = intro
				continue
			}
			out += line + "\n"
			continue
		}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Intro is finished when the comment end delimiter occurs.
The status afterwards is not defined. Comment/code pairs might follow,
or another multiline comment. Or the end of the file.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == intro {
			if isCommentEnd(line) {
				out += divEnd("intro doc")
				status = none
				continue
			}
			out += line + "\n"
			continue
		}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
A line comment can occur after code, after another line comment,
or when no other section is active.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == none || status == code {
			if isLineComment(line) {
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
If the last line was code, add a closing code fence.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
				if status == code {
					out += "```\n\n"
					out += divEnd("code")
					out += divEnd("ccpair")
					out += div("ccpair")
				}
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Multiline comments switch the status to none at the end.
In this case, start a new source section.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
				if status == none {
					out += div("source")
					out += div("ccpair")
				}
				status = comment
				out += div("comment")
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Strip the comment delimiters.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
				out += commentRe.ReplaceAllString(line, "") + "\n"
				continue
			}
		}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
While processing line comments.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == comment {
```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
If still looking at a line comment, strip the delims.
Else switch into code status.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
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

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
While processing code, look out for comments.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == code {

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
A line comment occurs. End the code section.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
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

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
A multiline comment starts. End the code section and switch to
single-column layout by closing the "source" div.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
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

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
At the end of a multiline comment, we don't know for sure
what comes next, so we set the status to none.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == doc {
			if isCommentEnd(line) {
				out += divEnd("doc")
				status = none
				continue
			}
			out += line + "\n"
			continue
		}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Outside any status? Just pass the line to the output.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
		if status == none {
			out += line + "\n"
		}
	}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
The last line in the file might be code.
We need a closing code fence then, and we need to close the divs, too.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	if status == code {
		out += "\n```\n"
		out += divEnd("code")
		out += divEnd("ccpair")
	}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Close the `gotohugo` namespace div.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	out += divEnd("gotohugo")

	return out
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
## Converting a file

### Again, some helper functions

`base` strips the extension from a filename. For some reason, this
function is missing from the standard path library.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func base(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
### Now the actual conversion

`convertFile` takes a file name, reads that file, converts it to
Markdown, and writes it to `*outDir/[post/]<basename>.md`
The path must already exist.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func convertFile(filename string) (err error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Cannot read file " + filename + "\n" + err.Error())
	}
	name := filepath.Base(filename)
	ext := ".md"
	basename := base(name) // strip ".go"
	outname := filepath.Join(*outDir, postDir, basename) + ext
	md := convert(string(src), basename)
	err = ioutil.WriteFile(outname, []byte(md), 0644) // -rw-r--r--
	if err != nil {
		return errors.Wrap(err, "Cannot write file "+outname)
	}
	return nil
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
newConvertFunc creates a function that converts the file described by `path`.
The function is used to create a `time.AfterFunc` function (which takes no parameters).
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func newConvertFunc(path string) func() {
	return func() {
		log.Println("Start converting   ", path+"...")
		err := convertFile(path)
		if err != nil {
			log.Println(err)
		}
		log.Println("Finished converting", path+".")
	}
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
`watchAndConvert` observes the file system under directory <dir>.
If a file named `<name>.go` in directory `<name>` has changed,
convert it to Hugo Markdown.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func watchAndConvert(dirname string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "Cannot create new Watcher")
	}
	defer watcher.Close()

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
A list of paths that shall trigger conversion. The key has the form "watch/watch.go".
After timer C times out, the path is sent through channel ch to `receivePathAndConvert()`.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	watchedPath := map[string]*time.Timer{}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Open the directory specified by -watch
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	dir, err := os.Open(dirname)
	if err != nil {
		return errors.Wrap(err, "Could not open "+dirname)
	}
	defer dir.Close()
	err = dir.Chdir()
	if err != nil {
		return errors.Wrap(err, "Could not cd to "+dirname)
	}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Read all entries in this directory
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	entries, err := dir.Readdir(0)
	if err != nil {
		return errors.Wrap(err, "Cannot read directory "+dirname)
	}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Watch the given directory.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	err = watcher.Add(dirname)
	if err != nil {
		return errors.Wrap(err, "Failed to add "+dirname+" to watcher")
	}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
If the entry is a directory, watch for creation of or changes to a
Go file under that dir of the same name as the dir, e.g. `watch/watch.go`.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	msg := ("Watching " + dirname + " and: ")

	for _, fsobj := range entries {
		if fsobj.IsDir() {
			fname := fsobj.Name()

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Watch the subdir for any changes.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
			err = watcher.Add(fname)
			if err != nil {
				return errors.Wrap(err, "Failed to add "+fname+" to watcher")
			}
			msg += fname + " "

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Remember the path that shall trigger conversion. As mentioned before,
this is a path like `watch/watch.go`.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
			fpath := filepath.Join(fname, fname+".go")
			dbg("Watching " + fpath + ".")
			watchedPath[fpath] = time.AfterFunc(time.Second, newConvertFunc(fpath))
			watchedPath[fpath].Stop()
		}
	}
	log.Println(msg)

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Avoid that deadlock detection kicks in.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	watchdog := time.NewTicker(10 * time.Second)

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Now look out for FS events.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	for {
		select {
		case event := <-watcher.Events:
			dbg("event:", event)
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				if watchedPath[event.Name] != nil {
					watchedPath[event.Name].Reset(time.Second) // Start if stopped. Reset if running.
				}
			}
		case err := <-watcher.Errors:
			return errors.Wrap(err, "Error while watching "+dirname)
		case <-watchdog.C:
			dbg("Watchdog triggered.")
		}
	}

	return nil
}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
## main - Where it all starts
{{% divend %}} <!--comment-->
{{% div code %}}

```go
func main() {

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
Start the Gops agent.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	if err := agent.Listen(nil); err != nil {
		log.Fatal(err)
	}

	flag.Parse()
	hugoDirEnv := os.Getenv("HUGODIR")

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
If $HUGODIR is set and -hugo isn't, copy $HUGODIR into *hugoDir.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	if len(*hugoDir) == 0 && len(hugoDirEnv) > 0 {
		*hugoDir = hugoDirEnv
	}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
If *hugoDir is set, use this instead of *outDir. Also set the subdirs accordingly.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	if len(*hugoDir) > 0 {
		*outDir = *hugoDir
		postDir = filepath.Join("content", "post")
		mediaDir = filepath.Join("static", "media") // media dir as Hugo sees it
		publicMediaDir = "media"                    // media dir as the Web server sees it
	}

```

{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% div ccpair %}}
{{% div comment %}}
With `-watch=<dir>`, watch the subdirs of `<dir>` for changes.
{{% divend %}} <!--comment-->
{{% div code %}}

```go
	if len(*watch) > 0 {
		log.Println("Running in watch mode. Hit Ctrl-C to stop.")
		err := watchAndConvert(*watch)
		if err != nil {
			log.Println(errors.Wrap(err, "Conversion Error"))
		}
	} else {
		for _, filename := range flag.Args() {
			log.Println("Converting", filename)
			err := convertFile(filename)
			if err != nil {
				log.Fatal(errors.Wrap(err, "Conversion Error"))
			}
		}
	}
	log.Println("Done.")
}


```
{{% divend %}} <!--code-->
{{% divend %}} <!--ccpair-->
{{% divend %}} <!--gotohugo-->
