
# helper tools
AWK ?= gawk
INSTALL ?= install
M4 ?= m4
SED ?= sed
EINFO ?= echo
PYTHON ?= python
CUT ?= cut

NAME ?= $(shell $(AWK) -F= '/^SELINUXTYPE/{ print $$2 }' /etc/selinux/config)
SHAREDIR ?= /usr/share/selinux
HEADERDIR ?= $(SHAREDIR)/$(NAME)/include

include $(HEADERDIR)/build.conf

# executables
PREFIX := /usr
BINDIR := $(PREFIX)/bin
SBINDIR := $(PREFIX)/sbin
CHECKMODULE := $(BINDIR)/checkmodule
SEMODULE := $(SBINDIR)/semodule
SEMOD_PKG := $(BINDIR)/semodule_package
XMLLINT := $(BINDIR)/xmllint

# set default build options if missing
TYPE ?= standard
DIRECT_INITRC ?= n
POLY ?= n
QUIET ?= y

genxml := $(PYTHON) $(HEADERDIR)/support/segenxml.py

docs := doc
polxml := $(docs)/policy.xml
xmldtd := $(HEADERDIR)/support/policy.dtd
metaxml := metadata.xml

globaltun = $(HEADERDIR)/global_tunables.xml
globalbool = $(HEADERDIR)/global_booleans.xml

# enable MLS if requested.
ifeq "$(TYPE)" "mls"
	M4PARAM += -D enable_mls
	CHECKPOLICY += -M
	CHECKMODULE += -M
endif

# enable MLS if MCS requested.
ifeq "$(TYPE)" "mcs"
	M4PARAM += -D enable_mcs
	CHECKPOLICY += -M
	CHECKMODULE += -M
endif

# enable distribution-specific policy
ifneq ($(DISTRO),)
	M4PARAM += -D distro_$(DISTRO)
endif

ifeq ($(DIRECT_INITRC),y)
	M4PARAM += -D direct_sysadm_daemon
endif

ifeq "$(UBAC)" "y"
	M4PARAM += -D enable_ubac
endif

# default MLS/MCS sensitivity and category settings.
MLS_SENS ?= 16
MLS_CATS ?= 256
MCS_CATS ?= 256

ifeq ($(QUIET),y)
	verbose := @
endif

M4PARAM += -D hide_broken_symptoms -D mls_num_sens=$(MLS_SENS) -D mls_num_cats=$(MLS_CATS) -D mcs_num_cats=$(MCS_CATS)

# policy headers
m4support = $(wildcard $(HEADERDIR)/support/*.spt)

header_layers := $(filter-out $(HEADERDIR)/support,$(shell find $(wildcard $(HEADERDIR)/*) -maxdepth 0 -type d))
header_xml := $(addsuffix .xml,$(header_layers))
header_interfaces := $(foreach layer,$(header_layers),$(wildcard $(layer)/*.if))

rolemap := $(HEADERDIR)/rolemap

local_layers := $(filter-out CVS tmp $(docs),$(shell find $(wildcard *) -maxdepth 0 -type d))
local_xml := $(addprefix tmp/, $(addsuffix .xml,$(local_layers)))

all_layer_names := $(sort $(notdir $(header_layers) $(local_layers)))

3rd_party_mods := $(wildcard *.te)
detected_mods := $(3rd_party_mods) $(foreach layer,$(local_layers),$(wildcard $(layer)/*.te))

detected_ifs := $(detected_mods:.te=.if)
detected_fcs := $(detected_mods:.te=.fc)
all_packages := $(notdir $(detected_mods:.te=.pp))

# figure out what modules we may want to reload
loaded_mods = $(addsuffix .pp,$(shell $(SEMODULE) -l | $(CUT) -f1))
sys_mods = $(wildcard $(SHAREDIR)/$(NAME)/*.pp)
match_sys = $(filter $(addprefix $(SHAREDIR)/$(NAME)/,$(loaded_mods)),$(sys_mods))
match_loc = $(filter $(all_packages),$(loaded_mods))

vpath %.te $(local_layers)
vpath %.if $(local_layers)
vpath %.fc $(local_layers)

########################################
#
# Functions
#

# parse-rolemap-compat modulename,outputfile
define parse-rolemap-compat
	$(verbose) $(M4) $(M4PARAM) $(rolemap) | \
		$(AWK) '/^[[:blank:]]*[A-Za-z]/{ print "gen_require(type " $$3 "; role " $$1 ";)\n$1_per_userdomain_template(" $$2 "," $$3 "," $$1 ")" }' >> $2
endef

# parse-rolemap modulename,outputfile
define parse-rolemap
	$(verbose) $(M4) $(M4PARAM) $(rolemap) | \
		$(AWK) '/^[[:blank:]]*[A-Za-z]/{ print "gen_require(type " $$3 "; role " $$1 ";)\n$1_per_role_template(" $$2 "," $$3 "," $$1 ")" }' >> $2
endef

# peruser-expansion modulename,outputfile
define peruser-expansion
	$(verbose) echo "ifdef(\`""$1""_per_role_template',\`" > $2
	$(call parse-rolemap,$1,$2)
	$(verbose) echo "')" >> $2

	$(verbose) echo "ifdef(\`""$1""_per_userdomain_template',\`" >> $2
	$(verbose) echo "errprint(\`Warning: per_userdomain_templates have been renamed to per_role_templates (""$1""_per_userdomain_template)'__endline__)" >> $2
	$(call parse-rolemap-compat,$1,$2)
	$(verbose) echo "')" >> $2
endef

.PHONY: clean all xml load reload
.SUFFIXES:
.SUFFIXES: .pp
# broken in make 3.81:
#.SECONDARY:

########################################
#
# Main targets
#

all: $(all_packages)

xml: $(polxml)

########################################
#
# Attempt to reinstall all installed packages
#
refresh:
	@$(EINFO) "Refreshing $(NAME) modules"
	$(verbose) $(SEMODULE) -b $(SHAREDIR)/$(NAME)/base.pp $(foreach mod,$(match_sys) $(match_loc),-i $(mod))

########################################
#
# Load module packages
#

load: tmp/loaded
tmp/loaded: $(all_packages)
	@$(EINFO) "Loading $(NAME) modules: $(basename $(notdir $?))"
	$(verbose) $(SEMODULE) $(foreach mod,$?,-i $(mod))
	@mkdir -p tmp
	@touch tmp/loaded

reload: $(all_packages)
	@$(EINFO) "Loading $(NAME) modules: $(basename $(notdir $^))"
	$(verbose) $(SEMODULE) $(foreach mod,$^,-i $(mod))
	@mkdir -p tmp
	@touch tmp/loaded

########################################
#
# Build module packages
#
tmp/%.mod: $(m4support) tmp/all_interfaces.conf %.te
	@$(EINFO) "Compiling $(NAME) $(basename $(@F)) module"
	@test -d $(@D) || mkdir -p $(@D)
	$(call peruser-expansion,$(basename $(@F)),$@.role)
	$(verbose) $(M4) $(M4PARAM) -s $^ $@.role > $(@:.mod=.tmp)
	$(verbose) $(CHECKMODULE) -m $(@:.mod=.tmp) -o $@

tmp/%.mod.fc: $(m4support) %.fc
	$(verbose) $(M4) $(M4PARAM) $^ > $@

%.pp: tmp/%.mod tmp/%.mod.fc
	@echo "Creating $(NAME) $(@F) policy package"
	$(verbose) $(SEMOD_PKG) -o $@ -m $< -f $<.fc

tmp/all_interfaces.conf: $(m4support) $(header_interfaces) $(detected_ifs)
	@test -d $(@D) || mkdir -p $(@D)
	@echo "ifdef(\`__if_error',\`m4exit(1)')" > tmp/iferror.m4
	@echo "divert(-1)" > $@
	$(verbose) $(M4) $^ tmp/iferror.m4 | sed -e s/dollarsstar/\$$\*/g >> $@
	@echo "divert" >> $@

# so users dont have to make empty .fc and .if files
$(detected_fcs):
	@touch $@
	
$(detected_ifs):
	@echo "## <summary>$(basename $(@D))</summary>" > $@

########################################
#
# Documentation generation
#
tmp/%.xml: %/*.te %/*.if
	@test -d $(@D) || mkdir -p $(@D)
	$(verbose) test -f $(HEADERDIR)/$*.xml || cat $*/$(metaxml) > $@
	$(verbose) $(genxml) -w -m $(sort $(basename $^)) >> $@

vars: $(local_xml)

$(polxml): $(header_xml) $(local_xml) $(globaltun) $(globalbool) $(detected_mods) $(detected_ifs)
	@echo "Creating $(@F)"
	@test -d $(@D) || mkdir -p $(@D)
	$(verbose) echo '<?xml version="1.0" encoding="ISO-8859-1" standalone="no"?>' > $@
	$(verbose) echo '<!DOCTYPE policy SYSTEM "$(notdir $(xmldtd))">' >> $@
	$(verbose) echo '<policy>' >> $@
	$(verbose) for i in $(all_layer_names); do \
		echo "<layer name=\"$$i\">" >> $@ ;\
		test -f $(HEADERDIR)/$$i.xml && cat $(HEADERDIR)/$$i.xml >> $@ ;\
		test -f tmp/$$i.xml && cat tmp/$$i.xml >> $@ ;\
		echo "</layer>" >> $@ ;\
	done
ifneq "$(strip $(3rd_party_mods))" ""
	$(verbose) echo "<layer name=\"third_party\">" >> $@
	$(verbose) echo "<summary>These are all third-party modules.</summary>" >> $@
	$(verbose) $(genxml) -w -m $(addprefix ./,$(basename $(3rd_party_mods))) >> $@
	$(verbose) echo "</layer>" >> $@
endif
	$(verbose) cat $(globaltun) $(globalbool) >> $@
	$(verbose) echo '</policy>' >> $@
	$(verbose) if test -x $(XMLLINT) && test -f $(xmldtd); then \
		$(XMLLINT) --noout --path $(dir $(xmldtd)) --dtdvalid $(xmldtd) $@ ;\
	fi

########################################
#
# Clean the environment
#

clean:
	rm -fR tmp
	rm -f *.pp
