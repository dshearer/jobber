<?php require("phplib/content-funcs.php"); ?>
<?php
$gSections = [
    "deployment" => [
        "title" => "Deployment",
        "page" => "deployment.html",
        "version" => "v1.2"
    ],
    "defining-jobs" => [
        "title" => "Defining Jobs",
        "sections" => [
            "overview" => [
                "title" => "Overview",
                "page" => "defining-jobs/overview.html",
                "version" => "v1.2"
            ],
            "section-prefs" => [
                "title" => "Preferences",
                "page" => "defining-jobs/prefs.html",
                "version" => "v1.2"
            ],
            "section-jobs" => [
                "title" => "Jobs",
                "page" => "defining-jobs/jobs.html",
                "version" => "v1.2"
            ],
            "error-handling" => [
                "title" => "Error-handling",
                "page" => "defining-jobs/error-handling.html",
                "version" => "v1.2"
            ],
        ]
    ],
    "loading-jobs" => [
        "title" => "Loading Jobs",
        "page" => "loading-jobs.html",
        "version" => "v1.1"
    ],
    "listing-jobs" => [
        "title" => "Listing Jobs",
        "page" => "listing-jobs.html",
        "version" => "v1.1"
    ],
    "listing-runs" => [
        "title" => "Listing Runs",
        "page" => "listing-runs.html",
        "version" => "v1.1"
    ],
    "testing-jobs" => [
        "title" => "Testing Jobs",
        "page" => "testing-jobs.html",
        "version" => "v1.1"
    ],
];
?>
<!DOCTYPE html>
<html lang="en">

<head>
<?php require("phplib/partials/head.html"); ?>

<link rel="stylesheet" href="/jobber/stylesheets/doc.css" />
<script src="/jobber/scripts/doc.js"></script>

<title>How to Use Jobber</title>

<script lang="test/javascript">
	<?php makeDocPageOnloadScript(); ?>
</script>
</head>

<body onload="onLoad()" data-spy="scroll" data-target="#toc-container"
  data-offset="100">

  <!-- NAV BAR -->
  <?php makeSubpageNavbar("doc"); ?>

  <header class="container">
    <h1>
      How to Use Jobber <br /> <small>Version
        <?php makeDocVersionSelect("1.2"); ?></small>
    </h1>
  </header>

  <!-- MAIN CONTENT -->
  <div id="main-container" class="container">
    <div class="row">

      <div class="col-md-3">
        <div id="toc-container" class="hidden-sm hidden-xs hidden-print"
          data-spy="affix" data-offset-top="150">
          <nav class="nav internal-nav">
            <ul class="nav-list-1">
              <li>
                <a class="top-nav-item nobr" href="#" target="_self">How
                  to Use</a>
                <ul class="nav-list-2">
                  <?php makeDocSectNav($gSections); ?>
                </ul>
              </li>
            </ul>
          </nav>
        </div>
      </div>
      <!-- col -->

      <section id="main-section" class="col-md-9 main" role="main">
        <?php makeDocSections($gSections); ?>
      </section>
    </div>
  </div>
  <!-- main content -->
</body>

</html>
