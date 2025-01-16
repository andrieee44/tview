# TVIEW

[NAME](#NAME)  
[SYNOPSIS](#SYNOPSIS)  
[DESCRIPTION](#DESCRIPTION)  
[OPTIONS](#OPTIONS)  
[CONFIGURATION](#CONFIGURATION)  
[EXAMPLES](#EXAMPLES)  
[EXAMPLE LF PREVIWER](#EXAMPLE%20LF%20PREVIWER)  
[ENVIRONMENT VARIABLES](#ENVIRONMENT%20VARIABLES)  
[SEE ALSO](#SEE%20ALSO)  
[AUTHOR](#AUTHOR)  

------------------------------------------------------------------------

## NAME <span id="NAME"></span>

tview − preview files in the terminal

## SYNOPSIS <span id="SYNOPSIS"></span>

**tview** \[**−−cache** *directory*\] \[**−−columns** *columns*\]
\[**−−config** *path*\] \[**−−rows** *rows*\] *FILE*

## DESCRIPTION <span id="DESCRIPTION"></span>

***tview*** is a file previewer for the terminal, originally designed
for *lf*(1). It runs external programs to preview file content based on
the file mimetype.

**tview** caches the previous preview and uses the cache if the contents
of the file did not change.

## OPTIONS <span id="OPTIONS"></span>

**−−cache** *directory*

The directory used to store the cached file previews. Defaults to
**"\${XDG_CACHE_HOME}/tview"**.

**−−columns** *columns*

The amount of terminal columns passed to the external programs with the
environment variable **"\$TVIEW_COLUMNS"**. Defaults to the column count
of the terminal **tview** is running in.

**−−config** *path*

The path to configuration file. Defaults to
**"\${XDG_CONFIG_HOME}/tview/config.json"**.

**−−rows** *rows*

The amount of terminal rows passed to the external programs with the
environment variable **"\$TVIEW_ROWS"**. Defaults to the row count of
the terminal **tview** is running in.

## CONFIGURATION <span id="CONFIGURATION"></span>

Configuration file is stored in
**"\${XDG_CONFIG_HOME}/tview/config.json"** by default. It is a JSON
file containing key−value entries of file mimetypes paired with a list
of external programs to be executed to generate previews. See the
**"EXAMPLES"** section for an example configuration snippet.

The external programs receive arguments from environmental variables set
during their execution. See the **"ENVIRONMENT VARIABLES"** section for
more information.

## EXAMPLES <span id="EXAMPLES"></span>

Preview a file:

**\$ tview ./myPhoto.png**

Example configuration snippet:

**{ "text/plain" : \[ "bat −−terminal−width "\$TVIEW_COLUMNS" −−
"\$TVIEW_FILE"" \] }**

## EXAMPLE LF PREVIWER <span id="EXAMPLE LF PREVIWER"></span>

In *lfrc*:

**set previewer ./tviewlf.sh**

In *tviewlf.sh*:

**\#! /bin/dash  
set −eu; tview −−columns "\$2" −−rows "\$3" −− "\$1"**

## ENVIRONMENT VARIABLES <span id="ENVIRONMENT VARIABLES"></span>

**\$TVIEW_FILE**

Path to the file that needs previewing.

**\$TVIEW_COLUMNS**

Amount of available terminal columns.

**\$TVIEW_ROWS**

Amount of available terminal rows.

## SEE ALSO <span id="SEE ALSO"></span>

*lf*(1)

## AUTHOR <span id="AUTHOR"></span>

andrieee44 (andrieee44@gmail.com)

------------------------------------------------------------------------
