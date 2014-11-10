#!/usr/bin/python

#  Author(s): Donald Miner <dminer@tresys.com>
#             Dave Sugar <dsugar@tresys.com>
#             Brian Williams <bwilliams@tresys.com>
#             Caleb Case <ccase@tresys.com>
#
# Copyright (C) 2005 - 2006 Tresys Technology, LLC
#      This program is free software; you can redistribute it and/or modify
#      it under the terms of the GNU General Public License as published by
#      the Free Software Foundation, version 2.

"""
	This script generates XML documentation information for layers specified
	by the user.
"""

import sys
import os
import glob
import re
import getopt

# GLOBALS

# Default values of command line arguments:
warn = False
meta = "metadata"
third_party = "third-party"
layers = {}
tunable_files = []
bool_files = []
xml_tunable_files = []
xml_bool_files = []
output_dir = ""

# Pre compiled regular expressions:

# Matches either an interface or a template declaration. Will give the tuple:
#	("interface" or "template", name)
# Some examples:
#	"interface(`kernel_read_system_state',`"
#	 -> ("interface", "kernel_read_system_state")
#	"template(`base_user_template',`"
#	 -> ("template", "base_user_template")
INTERFACE = re.compile("^\s*(interface|template)\(`(\w*)'")

# Matches either a gen_bool or a gen_tunable statement. Will give the tuple:
#	("tunable" or "bool", name, "true" or "false")
# Some examples:
#	"gen_bool(secure_mode, false)"
#	 -> ("bool", "secure_mode", "false")
#	"gen_tunable(allow_kerberos, false)"
#	 -> ("tunable", "allow_kerberos", "false")
BOOLEAN = re.compile("^\s*gen_(tunable|bool)\(\s*(\w*)\s*,\s*(true|false)\s*\)")

# Matches a XML comment in the policy, which is defined as any line starting
#  with two # and at least one character of white space. Will give the single
#  valued tuple:
#	("comment")
# Some Examples:
#	"## <summary>"
#	 -> ("<summary>")
#	"##		The domain allowed access.	"
#	 -> ("The domain allowed access.")
XML_COMMENT = re.compile("^##\s+(.*?)\s*$")


# FUNCTIONS
def getModuleXML(file_name):
	'''
	Returns the XML data for a module in a list, one line per list item.
	'''

	# Gather information.
	module_dir = os.path.dirname(file_name)
	module_name = os.path.basename(file_name)
	module_te = "%s/%s.te" % (module_dir, module_name)
	module_if = "%s/%s.if" % (module_dir, module_name)

	# Try to open the file, if it cant, just ignore it.
	try:
		module_file = open(module_if, "r")
		module_code = module_file.readlines()
		module_file.close()
	except:
		warning("cannot open file %s for read, skipping" % file_name)
		return []

	module_buf = []

	# Infer the module name, which is the base of the file name.
	module_buf.append("<module name=\"%s\" filename=\"%s\">\n" 
		% (os.path.splitext(os.path.split(file_name)[-1])[0], module_if))

	temp_buf = []
	interface = None

	# finding_header is a flag to denote whether we are still looking
	#  for the XML documentation at the head of the file.
	finding_header = True

	# Get rid of whitespace at top of file
	while(module_code and module_code[0].isspace()):
		module_code = module_code[1:]

	# Go line by line and figure out what to do with it.
	line_num = 0
	for line in module_code:
		line_num += 1
		if finding_header:
			# If there is a XML comment, add it to the temp buffer.
			comment = XML_COMMENT.match(line)
			if comment:
				temp_buf.append(comment.group(1) + "\n")
				continue

			# Once a line that is not an XML comment is reached,
			#  either put the XML out to module buffer as the
			#  module's documentation, or attribute it to an
			#  interface/template.
			elif temp_buf:
				finding_header = False
				interface = INTERFACE.match(line)
				if not interface:
					module_buf += temp_buf
					temp_buf = []
					continue

		# Skip over empty lines
		if line.isspace():
			continue

		# Grab a comment and add it to the temprorary buffer, if it
		#  is there.
		comment = XML_COMMENT.match(line)
		if comment:
			temp_buf.append(comment.group(1) + "\n")
			continue

		# Grab the interface information. This is only not true when
		#  the interface is at the top of the file and there is no
		#  documentation for the module.
		if not interface:
			interface = INTERFACE.match(line)
		if interface:
			# Add the opening tag for the interface/template
			groups = interface.groups()
			module_buf.append("<%s name=\"%s\" lineno=\"%s\">\n" % (groups[0], groups[1], line_num))

			# Add all the comments attributed to this interface to
			#  the module buffer.
			if temp_buf:
				module_buf += temp_buf
				temp_buf = []

			# Add default summaries and parameters so that the
			#  DTD is happy.
			else:
				warning ("unable to find XML for %s %s()" % (groups[0], groups[1]))	
				module_buf.append("<summary>\n")
				module_buf.append("Summary is missing!\n")
				module_buf.append("</summary>\n")
				module_buf.append("<param name=\"?\">\n")
				module_buf.append("<summary>\n")
				module_buf.append("Parameter descriptions are missing!\n")
				module_buf.append("</summary>\n")
				module_buf.append("</param>\n")

			# Close the interface/template tag.
			module_buf.append("</%s>\n" % interface.group(1))

			interface = None
			continue



	# If the file just had a header, add the comments to the module buffer.
	if finding_header:
		module_buf += temp_buf
	# Otherwise there are some lingering XML comments at the bottom, warn
	#  the user.
	elif temp_buf:
		warning("orphan XML comments at bottom of file %s" % file_name)

	# Process the TE file if it exists.
	module_buf = module_buf + getTunableXML(module_te, "both")

	module_buf.append("</module>\n")

	return module_buf

def getTunableXML(file_name, kind):
	'''
	Return all the XML for the tunables/bools in the file specified.
	'''

	# Try to open the file, if it cant, just ignore it.
	try:
		tunable_file = open(file_name, "r")
		tunable_code = tunable_file.readlines()
		tunable_file.close()
	except:
		warning("cannot open file %s for read, skipping" % file_name)
		return []

	tunable_buf = []
	temp_buf = []

	# Find tunables and booleans line by line and use the comments above
	# them.
	for line in tunable_code:
		# If it is an XML comment, add it to the buffer and go on.
		comment = XML_COMMENT.match(line)
		if comment:
			temp_buf.append(comment.group(1) + "\n")
			continue

		# Get the boolean/tunable data.
		boolean = BOOLEAN.match(line)

		# If we reach a boolean/tunable declaration, attribute all XML
		#  in the temp buffer to it and add XML to the tunable buffer.
		if boolean:
			# If there is a gen_bool in a tunable file or a
			# gen_tunable in a boolean file, error and exit.
			# Skip if both kinds are valid.
			if kind != "both":
				if boolean.group(1) != kind:
					error("%s in a %s file." % (boolean.group(1), kind))

			tunable_buf.append("<%s name=\"%s\" dftval=\"%s\">\n" % boolean.groups())
			tunable_buf += temp_buf
			temp_buf = []
			tunable_buf.append("</%s>\n" % boolean.group(1))

	# If there are XML comments at the end of the file, they arn't
	# attributed to anything. These are ignored.
	if len(temp_buf):
		warning("orphan XML comments at bottom of file %s" % file_name)


	# If the caller requested a the global_tunables and global_booleans to be
	# output to a file output them now
	if len(output_dir) > 0:
		xmlfile = os.path.split(file_name)[1] + ".xml"

		try:
			xml_outfile = open(output_dir + "/" + xmlfile, "w")
			for tunable_line in tunable_buf:
				xml_outfile.write (tunable_line)
			xml_outfile.close()
		except:
			warning ("cannot write to file %s, skipping creation" % xmlfile)

	return tunable_buf

def getXMLFileContents (file_name):
	'''
	Return all the XML in the file specified.
	'''

	tunable_buf = []
	# Try to open the xml file for this type of file
	# append the contents to the buffer.
	try:
		tunable_xml = open(file_name, "r")
		tunable_buf += tunable_xml.readlines()
		tunable_xml.close()
	except:
		warning("cannot open file %s for read, assuming no data" % file_name)

	return tunable_buf

def getPolicyXML():
	'''
	Return the compelete reference policy XML documentation through a list,
	one line per item.
	'''

	policy_buf = []
	policy_buf.append("<policy>\n")

	# Add to the XML each layer specified by the user.
	for layer in layers.keys ():
		policy_buf += getLayerXML(layer, layers[layer])

	# Add to the XML each tunable file specified by the user.
	for tunable_file in tunable_files:
		policy_buf += getTunableXML(tunable_file, "tunable")

	# Add to the XML each XML tunable file specified by the user.
	for tunable_file in xml_tunable_files:
		policy_buf += getXMLFileContents (tunable_file)

	# Add to the XML each bool file specified by the user.
	for bool_file in bool_files:
		policy_buf += getTunableXML(bool_file, "bool")

	# Add to the XML each XML bool file specified by the user.
	for bool_file in xml_bool_files:
		policy_buf += getXMLFileContents (bool_file)

	policy_buf.append("</policy>\n")

	return policy_buf

def usage():
	"""
	Displays a message describing the proper usage of this script.
	"""

	sys.stdout.write("usage: %s [-w] [-mtb] <file>\n\n" % sys.argv[0])
	sys.stdout.write("-w --warn\t\t\tshow warnings\n"+\
	"-m --module <file>\t\tname of module to process\n"+\
	"-t --tunable <file>\t\tname of global tunable file to process\n"+\
	"-b --boolean <file>\t\tname of global boolean file to process\n\n")

	sys.stdout.write("examples:\n")
	sys.stdout.write("> %s -w -m policy/modules/apache\n" % sys.argv[0])
	sys.stdout.write("> %s -t policy/global_tunables\n" % sys.argv[0])

def warning(description):
	'''
	Warns the user of a non-critical error.
	'''

	if warn:
		sys.stderr.write("%s: " % sys.argv[0] )
		sys.stderr.write("warning: " + description + "\n")

def error(description):
	'''
	Describes an error and exists the program.
	'''

	sys.stderr.write("%s: " % sys.argv[0] )
        sys.stderr.write("error: " + description + "\n")
        sys.stderr.flush()
        sys.exit(1)



# MAIN PROGRAM

# Defaults
warn = False
module = False
tunable = False
boolean = False

# Check that there are command line arguments.
if len(sys.argv) <= 1:
	usage()
	sys.exit(1)

# Parse command line args
try:
	opts, args = getopt.getopt(sys.argv[1:], 'whm:t:b:', ['warn', 'help', 'module=', 'tunable=', 'boolean='])
except getopt.GetoptError:
	usage()
	sys.exit(2)
for o, a in opts:
	if o in ('-w', '--warn'):
		warn = True
	elif o in ('-h', '--help'):
		usage()
		sys.exit(0)
	elif o in ('-m', '--module'):
		module = a
		break
	elif o in ('-t', '--tunable'):
		tunable = a
		break
	elif o in ('-b', '--boolean'):
		boolean = a
		break
	else:
		usage()
		sys.exit(2)

if module:
	sys.stdout.writelines(getModuleXML(module))
elif tunable:
	sys.stdout.writelines(getTunableXML(tunable, "tunable"))
elif boolean:
	sys.stdout.writelines(getTunableXML(boolean, "bool"))
else:
	usage()
	sys.exit(2)

