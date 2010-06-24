include $(GOROOT)/src/Make.$(GOARCH)

TARG=luapatterns
GOFILES =\
		 classes.go\
		 luapatterns.go\

include $(GOROOT)/src/Make.pkg
