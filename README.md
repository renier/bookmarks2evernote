# bookmarks2evernote

Reads an exported html file of bookmarks to create an Evernote file (.enex)
that can be imported into Evernote. A note is created for each bookmark. It is best to import
the resulting Evernote file into its own notebook. Tags and bookmark description are preserved along
with the title and url of the bookmark. This program currently works best when using an export file from
Delicious, or otherwise when the list of bookmarks is a flat list (no folders).

## Install

    go get github.com/renier/bookmarks2evernote

## Usage

    $ bookmarks2evernote <input file> <output file>

Import the resulting file into evernote.

### Example

    $ bookmarks2evernote mybookmarks.html notebook.enex

