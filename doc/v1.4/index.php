<?php require("phplib/content-funcs.php"); ?>
<?php
$gSections = [
    "deployment" => [
        "title" => "Deployment",
        "page" => "deployment.html",
        "version" => "v1.4"
    ],
    "jobfile" => [
        "title" => "Putting Jobber to Work",
        "sections" => [
            "overview" => [
                "title" => "Overview",
                "page" => "jobfile/overview.html",
                "version" => "v1.4"
            ],
            "specifying-commands" => [
              "title" => "Specifying Commands",
              "page" => "jobfile/specifying-commands.html",
              "version" => "v1.4"
            ],
            "time-strings" => [
                "title" => "Time strings",
                "page" => "jobfile/time-strings.html",
                "version" => "v1.3"
            ],
            "error-handling" => [
                "title" => "Error-handling",
                "page" => "jobfile/error-handling.html",
                "version" => "v1.4"
            ],
            "notifications" => [
                "title" => "Result Sinks",
                "page" => "jobfile/result-sinks.html",
                "version" => "v1.4"
            ],
            "run-history" => [
                "title" => "Keeping a Log of Job Runs",
                "page" => "jobfile/run-log.html",
                "version" => "v1.4"
            ],
        ],
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
    "pausing-resuming-jobs" => [
        "title" => "Pausing and Resuming Jobs",
        "page" => "pausing-resuming-jobs.html",
        "version" => "v1.2"
    ],
    "cat-jobs" => [
        "title" => "Printing a Job's Command",
        "page" => "cat-jobs.html",
        "version" => "v1.2"
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
        <?php makeDocVersionSelect("1.4"); ?></small>
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
              <li class="top-nav-item nobr" target="_self"><a href="#">How
                  to Use</a></li>
              <li>
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
