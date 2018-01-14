var g_versions = [ "1.3", "1.2", "1.1" ];

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
		window.location.pathname = "/jobber/doc/v" + version + "/";
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

function _makeDocNav(navUl, sections) {
	for ( var sectId in sections) {
		var section = sections[sectId];
		
		// make nav bar item
		var li = $("<li class=\"nobr\"></li>").appendTo(navUl);
		var a = $("<a target=\"_self\"></a>").appendTo(li);
		a.attr("href", "#" + sectId);
		a.text(section.title);
		
		if (section.sections) {
			// add items for subsections
			var subUl = $("<ul class=\"nav-list-3\"></ul>").appendTo(li);
			for (var subSectId in section.sections)
			{
				var subSection = section.sections[subSectId];
				
				// make nav bar item
				var subLi = $("<li class=\"nobr\"></li>").appendTo(subUl);
				var subA = $("<a target=\"_self\"></a>").appendTo(subLi);
				subA.attr("href", "#" + subSectId);
				subA.text(subSection.title);
			}
		}
	}
}

function _makeSections(sectionContainer, sections) {
	for ( var sectId in sections) {
		var section = sections[sectId];

		if (section.sections) {
			// make section
			var div = $("<section></section>").appendTo(sectionContainer);
			div.attr("id", sectId);
			var header = $("<h2></h2>").appendTo(div);
			header.text(section.title);
			
			// load subsections
			for (var subSectId in section.sections)
			{
				var subSection = section.sections[subSectId];
				var subDiv = $("<div></div>").appendTo(div);
				subDiv.attr("id", subSectId);
				subDiv.load("/jobber/doc/" + subSection.version + 
						"/partials/" + subSection.page);
			}
		}
		else {
			// load section
			var div = $("<div></div>").appendTo(sectionContainer);
			div.attr("id", sectId);
			div.load("/jobber/doc/" + section.version + "/partials/" + 
					section.page);
		}
	}
}

function addSections(navUl, sectionContainer, sections) {
	_makeDocNav(navUl, sections);
	_makeSections(sectionContainer, sections);
}
