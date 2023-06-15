# Project Manager

Creative name, right?  This is a simple cli util I've made to manage some of my projects &
their corresponding envs on my systems.  This project mostly just covers my needs and might
not be ideal for everyone else.

## Features

- configurable `env` file which applies to the shell instance opened w/ project
- custom templates for new projects (builtin templates for go binaries, go libraries and ts npm packages)
- easy navigation from project to project
- simple project creation
- shell completion support

## Setup

This is mostly made for personal usage so setup isn't streamlined (though if requested I'd be happy to make
something for that).  

1. go install github.com/AllAwesome497/pmg
2. Set `PROJMGR_PROJDIR` to a directory in your `.profile` (if the directory doesn't exist, pmg should create it)
3. (optional) Set `PROJMGR_TEMPLATE_PATH` to a directory containing templates. Create this directory & set env var
   in `.profile` to start creating custom templates w/ `pmg create-template`.

To generate completion, follow the steps provided via `pmg completion ${YOUR_SHELL_NAME} --help`. (\*with zsh I have
a directory in ~/.local/completions which I append to fpath; this is where I store custom completions instead of fpath[1]).

## .pmg directory

This directory contains special files for the project manager. Currently there's only 2 of these:
- `.pmg/setup` - should be an executable file.  Only exists in templates.  When a project is created from a template,
  this file will be executed in the working directory of the new project with arg1 equal to the project name.  This 
  is deleted from the project's files once executed.
- `.pmg/env` - environment variables to be injected into shell when this project is opened by pmg.  The values of
  these env vars can reference other env vars.  For examples see cmd/templates.  Note that env vars set in .profile
  will overwrite these values, to avoid this you may want to have `.profile` do something like `MY_ENV_VAR=${MY_ENV_VAR:=value}`
  (i.e.: don't overwrite the env var if it already exists)

## Contributing

If you want to work on this package, you can import it via `pmg i https://github.com/AllAwesome497/pmg`.  This project
has some special env vars to map templates to cmd/templates (source for builtin templates) and to change the project dir
to a local directory (to make template testing easier / more contained).

## To-dos / might do

will add if someone asks or if i need it:

- more templates
- more docs 
- smarter logic for default project directory 
- project-specific shell history (tbd how this would work - possibly write and read 
  from both but prioritize project files when reading)
