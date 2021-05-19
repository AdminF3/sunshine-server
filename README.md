# sunshine

Backend server for the Sunshine platform.

[*API docs*](./doc/README.md)

## Requirements

- Go 1.10+
- PostgreSQL 10+
- Set `$GOPATH`
- Put `$GOPATH/bin` inside `$PATH` (e.g. `export PATH=$PATH:$GOPATH/bin`)
- User with superuser access to the running PostgreSQL and do `createdb sunshine`

## Development setup

1. Install `texlive` and Libertine (in Arch the packages are `texlive-bin`,
   `texlive-most`, `texlive-lang` and `ttf-linux-libertine`, in Ubuntu -
   `texlive-xetex`, `texlive-fonts-extra`, `texlive-latex-extra` and
   `fonts-linuxlibertine`)
2. Run `make build migrate`
3. Profit!

## Running the tests

	make test

## Running the application

1. Make sure you've executed all steps start from `2` from "Development setup" after each `git pull`.
2. Run `sunshine serve`
3. Profit!
