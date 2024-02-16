# Installation vars
ifndef PREFIX
PREFIX=/usr/local
endif
ifndef DATADIR
DATADIR=$(PREFIX)/share
endif

# Go build shim
revolver:
	go build
.PHONY: revolver

install: revolver README.md LICENSE
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m755 revolver $(DESTDIR)$(PREFIX)/bin/revolver
	install -d $(DESTDIR)$(DATADIR)/doc/revolver
	install -m644 README.md $(DESTDIR)$(DATADIR)/doc/revolver/README.md
	install -d $(DESTDIR)$(DATADIR)/licenses/revolver
	install -m644 LICENSE $(DESTDIR)$(DATADIR)/licenses/revolver/LICENSE
.PHONY: install

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/revolver
	rm -rf $(DESTDIR)$(DATADIR)/doc/revolver
	rm -rf $(DESTDIR)$(DATADIR)/licenses/revolver
.PHONY: uninstall
