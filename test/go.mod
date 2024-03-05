module github.com/alcomist/go-portfolio/test

go 1.19

replace github.com/alcomist/go-portfolio/internal => ./../internal

replace github.com/alcomist/go-portfolio/task => ./../task

require github.com/alcomist/go-portfolio/internal v0.0.0-00010101000000-000000000000

require github.com/google/uuid v1.6.0 // indirect
