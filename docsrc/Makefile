
OUTDIR = $(CURDIR)/../docs

all: src/node_modules
	make clean
	cd $(CURDIR)/src && \
		gitbook build && \
		mv $(CURDIR)/src/_book $(OUTDIR)

src/node_modules:
	cd $(CURDIR)/src && \
		yarn install && \
		gitbook install

clean:
	rm -rf $(CURDIR)/src/assets/images/uml/*
	rm -rf $(OUTDIR)
