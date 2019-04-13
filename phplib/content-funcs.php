<?php

$DOC_VERSIONS = ["1.1", "1.2", "1.3", "1.4"];
$DOC_DEFAULT_VERSION_IDX = 3;

function makeSubpageNavbar($currSubpage)
{
    global $DOC_VERSIONS;
    global $DOC_DEFAULT_VERSION_IDX;

    $subpages = [
        "download" => [
            "uri" => "/jobber/download/",
            "title" => "Download"
        ],
        "doc" => [
            "uri" => "/jobber/doc/v{$DOC_VERSIONS[$DOC_DEFAULT_VERSION_IDX]}/",
            "title" => "How to Use"
        ],
        "security" => [
            "uri" => "/jobber/security/",
            "title" => "Security"
        ],
        "blog" => [
            "uri" => "/jobber/blog/",
            "title" => "Blog"
        ]
    ];
    $subpageOrder = ["download", "doc", "security", "blog"];

    ?>
      <nav class="navbar navbar-default">
        <div class="container">
          <div class="navbar-header">
            <button type="button" class="navbar-toggle collapsed"
              data-toggle="collapse" data-target="#navbar-collapse"
              aria-expanded="false">
              <span class="sr-only">Toggle navigation</span> <span class="icon-bar"></span>
              <span class="icon-bar"></span> <span class="icon-bar"></span>
            </button>
            <a class="navbar-brand" href="/jobber/">Jobber</a>
          </div>
          <div class="collapse navbar-collapse" id="navbar-collapse">
            <ul class="nav navbar-nav">
            <?php
            foreach ($subpageOrder as $spId)
            {
                $subpage = $subpages[$spId];
                if ($spId == $currSubpage) {
                    ?><li class="active"><?php
                }
                else {
                    ?><li><?php
                }
                ?><a href="<?= $subpage["uri"] ?>">
                    <?= $subpage["title"] ?>
                  </a>
                </li>
                <?php
            }
            ?>
            </ul>
          </div>
        </div>
      </nav>
    <?php
}

function makeDocPageOnloadScript()
{
  ?>
  function onLoad() {
		// reset version selector
    var select = $("header h1 small select");
		var opt = select.find("option[selected]")[0];
    select.val(opt.value);
	}
  <?php
}

function makeDocVersionSelect($currVersion)
{
  global $DOC_VERSIONS;
  $revVersions = array_reverse($DOC_VERSIONS);

  $onChangeJs =
    "var ver = event.target.value; " .
    "if (ver == '".$currVersion."') { return; } " .
    "window.location.pathname = '/jobber/doc/v' + ver + '/';";

  ?><select onchange="<?= $onChangeJs ?>"><?php
    foreach ($revVersions as $ver)
    {
      $selected = ($ver == $currVersion) ? "selected" : "";
      ?><option value="<?= $ver ?>" <?= $selected ?> ><?= $ver ?></option><?php
    }
  ?></select><?php
}

function makeDocSections($sections)
{
    foreach ($sections as $sectId => $sect)
    {
        if (array_key_exists("sections", $sect)) {
            // make section
            ?>
            <section id="<?= $sectId ?>">
              <h2><?= $sect["title"] ?></h2>
            <?php

            // load subsections
            foreach ($sect["sections"] as $subsectId => $subsect)
            {
                ?><div id="<?= $subsectId ?>"><?php
                    require("doc/{$subsect['version']}/" .
                        "partials/{$subsect['page']}");
                ?></div><?php
            }

            ?></section><?php
        }
        else {
            // load section
            ?><div id="<?= $sectId ?>"><?php
                require("doc/{$sect['version']}/" .
                    "partials/{$sect['page']}");
            ?></div><?php
        }
    }
}

function makeDocSectNav($sections) {
    foreach ($sections as $sectId => $sect)
    {
        ?>
        <li class="nobr">
          <a target="_self" href="#<?= $sectId ?>">
            <?= $sect["title"] ?>
          </a>
        <?php

        if (array_key_exists("sections", $sect)) {
            ?><ul class="nav-list-3"><?php
            foreach ($sect["sections"] as $subsectId => $subsect)
            {
                ?>
                <li class="nobr">
                  <a target="_self" href="#<?= $subsectId ?>">
                    <?= $subsect["title"] ?>
                  </a>
                </li>
                <?php
            }
            ?></ul><?php
        }

        ?></li><?php
    }
}

?>
