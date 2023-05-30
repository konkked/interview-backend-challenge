# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean

# Output directory
OUTDIR = out

# Build target
build:
	@mkdir -p $(OUTDIR)
	$(GOBUILD) -o $(OUTDIR) ./...

# Clean target
clean:
	$(GOCLEAN)
	rm -rf $(OUTDIR)
