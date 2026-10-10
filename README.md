This is a general LSP for Mecrisp Forth. Currently it supports Hover over which will show the word the stack effects and a short description if available.
It also supports completion for words in the default dictionary.

Whats coming... 

* Adding Hover and completion for user defined words in the current file
* Builds for Linux and other platforms

To install you would add it to your editors LSP servers.

For Sublimetext this is...

```
// Settings in here override those in "LSP/LSP.sublime-settings"
{

  // Language server configurations
  "clients": {
    "goforth": {
      // enable this configuration
      "enabled": false,
      // the startup command -- what you would type in a terminal, this path will vary for each system
      "command": ["/home/morris/Stuff/golang/github.com/wolfmanjm/wjm_forth_lsp/wjm_forth_lsp"],
      // the selector that selects which type of buffers this language server attaches to
      "selector": "source.forth",
    },
  },
}
```
And for each project the LSP needs to be enabled...

```
{
	"settings":
	{
		"LSP":
		{
			"goforth":
			{
				"enabled": true,
			},
		},
	},
}

```

