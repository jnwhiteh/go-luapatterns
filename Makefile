include $(GOROOT)/src/Make.$(GOARCH)

TARG=luapatterns
GOFILES =\
		 luapatterns.go\

include $(GOROOT)/src/Make.pkg
