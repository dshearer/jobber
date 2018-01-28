<?php 

function makeSubpageNavbar($currSubpage)
{
    $subpages = [
        "download" => [
            "uri" => "/jobber/download/",
            "title" => "Download"
        ],
        "doc" => [
            "uri" => "/jobber/doc/v1.2/",
            "title" => "How to Use"
        ],
        "security" => [
            "uri" => "/jobber/security/",
            "title" => "Security"
        ],
        "blog" => [
            "uri" => "/jobber/blog/2018/01/14/go-review/",
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
                    ?><li class="active">
                        <a href="#"><?= $subpage["title"] ?></a>
                      </li>
                    <?php
                }
                else {
                    ?><li>
                        <a href="<?= $subpage["uri"] ?>"><?= $subpage["title"] ?></a>
                      </li>
                    <?php
                }
            }
            ?>
            </ul>
          </div>
        </div>
      </nav>
    <?php
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