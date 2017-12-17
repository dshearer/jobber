var g_versions = [ "1.2", "1.1" ];

/*
 * Make a select control that lists the documented versions and lets the user
 * switch to the documentation for other verions.
 */
function makeVersionsSelect(currVersion) {
	// make select element
	var select = $("<select></select>");

	// add onchange handler
	select.on("change", function() {
		// get selected version
		var opts = select.prop("options");
		var version = opts[opts.selectedIndex].text;

		// go to documentation for that version
		window.location.pathname = "/jobber/doc/v" + version;
	});

	// add option elements
	var selectedOptIndex = undefined;
	for ( var idx in g_versions) {
		var version = g_versions[idx];
		var option = $("<option></option>");
		select.append(option);
		if (version == currVersion) {
			selectedOptIndex = idx;
		}
		option.text(version);
	}

	// select current version
	if (selectedOptIndex !== undefined) {
		select.prop("selectedIndex", selectedOptIndex);
	}

	return select;
}

function addSections(navUl, sectionContainer, sections) {
	for ( var sectId in sections) {
		// make nav bar items
		var section = sections[sectId];
		var li = $("<li class=\"nobr\"></li>").appendTo(navUl);
		var a = $("<a target=\"_self\"></a>").appendTo(li);
		a.attr("href", "#" + sectId);
		a.text(section.title);

		// load section
		var div = $("<div></div>").appendTo(sectionContainer);
		div.attr("id", sectId);
		div.load("/jobber/doc/" + section.version + "/partials/" + 
				section.page);
	}
}
