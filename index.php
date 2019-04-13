<?php
require("phplib/releases.php");
require("phplib/content-funcs.php");
$release = latestRelease();
?>
<!DOCTYPE html>
<html lang="en">
<head>
<?php require("phplib/partials/head.html"); ?>

<title>Jobber: An alternative to cron, with sophisticated
  status-reporting and error-handling.</title>
</head>

<body itemscope itemtype="http://schema.org/SoftwareApplication">

  <!-- HEADER -->
  <header class="container">
    <div class="pull-left">
      <h1 itemprop="name">Jobber</h1>
    </div>

    <div class="banner-btn hidden-print pull-right"
      style="margin-top: 1em; margin-bottom: 1em;">
      <a href="/jobber/download/" class="btn btn-lg btn-default">
        <div class="fa fa-arrow-circle-o-down fa-2x pull-left"
          aria-hidden="true"></div>
        <div class="pull-right">
          Get Jobber <span class="banner-btn-details"><?= $release["name"] ?> | <?= $release["date"] ?></span>
        </div>
      </a>
    </div>
  </header>

  <!-- NAV BAR -->
  <nav class="navbar navbar-default">
    <div class="container">
      <ul class="nav navbar-nav">
        <li><a href="/jobber/download/">Download</a></li>
        <li><a href="<?= "/jobber/doc/v{$DOC_VERSIONS[$DOC_DEFAULT_VERSION_IDX]}/" ?>">How to Use</a></li>
        <li><a href="/jobber/security/">Security</a></li>
        <li><a href="/jobber/blog/">Blog</a></li>
      </ul>
      <ul class="nav navbar-nav navbar-right github-link">
        <li><a class="github-link"
          href="https://github.com/dshearer/jobber"> <span
            class="fa fa-github fa-lg" aria-hidden="true"></span> View on GitHub
            <span class="fa fa-external-link" aria-hidden="true"></span>
        </a></li>
      </ul>
    </div>
  </nav>

  <main class="container" role="main">

    <!-- Intro -->
    <section id="intro">
      <p class="lead" itemprop="description">
        Jobber is a <span itemprop="applicationCategory"> utility</span> for
        Unix-like systems that can run arbitrary commands, or
        &ldquo;jobs&rdquo;, according to a schedule. It is meant to be a
        better alternative to the classic Unix utility <a
          href="https://en.wikipedia.org/wiki/Cron">cron</a>.
      </p>

      <p>Along with the functionality of cron, Jobber also provides:</p>

      <ul itemprop="featureList">
        <li><strong>Job execution history</strong>: you can see what jobs
          have recently run, and whether they succeeded or failed.</li>
        <li><strong>Sophisticated error handling</strong>: you can control
          whether and when a job is run again after it fails. For example, after
          an initial failure of a job, Jobber can schedule future runs using an
          exponential backoff algorithm.</li>
        <li><strong>Sophisticated error reporting</strong>: you can control
          whether Jobber notifies you about each failed run, or only about jobs
          that have been disabled due to repeated failures.</li>
      </ul>
    </section>

    <!-- News -->
    <section>
      <h2>Recent Blog Posts</h2>
      <ul>
        <li class="h4">
          <a href="https://dev.to/dshearer/restful-security-plug-the-leaks-npa">
            RESTful Security: Plug the Leaks! <span class="fa fa-external-link"></span>
          </a>
          <small>7 Apr 2018</small>
        </li>
        <li class="h4">
          <a href="https://dev.to/dshearer/how-to-support-multiple-oses-with-one-mac-4145">
          How to Support Multiple OSes with One Mac <span class="fa fa-external-link"></span>
          </a>
          <small>16 Jan 2018</small>
        </li>
      </ul>
    </section>

    <!-- Project Status -->
    <section id="project-status">
      <h2>Project Status</h2>

      <p>
        Jobber is stable. <a href="https://github.com/dshearer/jobber/releases">Production
          releases</a> have been made.
      </p>

      <p>
        Bug reports and feature requests are welcome; please report them on <a
          href="https://github.com/dshearer/jobber/issues">Jobber&rsquo;s
          GitHub site</a>.
      </p>

      <p>
        Please post more open questions or comments on <a
          href="https://groups.google.com/d/forum/jobber-proj">the mailing
          list</a>.
      </p>

      <p>
        <strong>Contributions are welcome as well!</strong>
      </p>
    </section>
  </main>

  <!-- FOOTER  -->
  <footer class="small">
    <p>Copyright &#0169; 2014&ndash;2018 C.&nbsp;Dylan Shearer</p>
  </footer>
</body>
</html>
