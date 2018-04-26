P = go
DEST = $(HOME)/.codeblocks/share/codeblocks/plugins/GoTools/gotools.server.codeblocks

Debug: run
Release: build

server: $(P)

#EXPS = export GOPATH=$(HOME)/WORK/cf-c/go; export PATH=$(PATH):$(HOME)/go/bin; 

run: $(P)
	./$< <../go-tests/test.cmds
#	./$< <in.in

clean:
	$(EXPS) go clean

$(P)::
	$(EXPS) go build $(GOPATH)/src/github.com/kouzdra/go-analyzer/$(P).go

commit:
	git commit -a
