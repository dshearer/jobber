<?php
require("phplib/content-funcs.php");
require("phplib/releases.php");
$release = latestRelease();
?>
<!DOCTYPE html>
<html lang="en">

<head>
<?php require("phplib/partials/head.html"); ?>

<title>Download Jobber</title>
</head>

<body data-spy="scroll" data-target="#toc-container" data-offset="100">

  <!-- NAV BAR -->
  <?php makeSubpageNavbar("download"); ?>

  <header class="container">
    <h1>Download Jobber</h1>
  </header>

  <!-- MAIN CONTENT -->
  <div class="container">

    <section id="license">
      <h2>License</h2>

      <p>
        Jobber's source code is copyright &#0169; <span itemprop="copyrightYear">
          2014&ndash;2020</span> <span itemprop="copyrightHolder" itemscope
          itemtype="http://schema.org/Person"><span itemprop="name">C.&nbsp;Dylan
          Shearer</span></span>. It is licensed according to the MIT License:
      </p>

      <div itemprop="license" itemscope
        itemtype="http://schema.org/CreativeWork">
        <blockquote class="license" itemprop="text">
          <p>Permission is hereby granted, free of charge, to any person
            obtaining a copy of this software and associated documentation files
            (the "Software"), to deal in the Software without restriction,
            including without limitation the rights to use, copy, modify, merge,
            publish, distribute, sublicense, and/or sell copies of the Software,
            and to permit persons to whom the Software is furnished to do so,
            subject to the following conditions:</p>

          <p>The above copyright notice and this permission notice shall be
            included in all copies or substantial portions of the Software.</p>

          <p>THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
            EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
            MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
            NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
            BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
            ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
            CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
            SOFTWARE.</p>
        </blockquote>
      </div>
    </section>

    <!-- Downloading -->
    <section id="downloading">
      <h2>Downloading</h2>

      <p>The latest release is <?= $release["name"] ?>, which was made
      on <?= $release["date"] ?>.</p>
      
      <p><a href="<?= $release["rel_notes_url"] ?>">Release notes.</a>
      
      <h3>Source</h3>
      <ul>
          <?php foreach (["tarball", "zipball"] as $src) { ?>
          <li>
          <a href="<?= $release["{$src}_url"] ?>">
            <span class="fa fa-download" aria-hidden="true">&nbsp;</span><?= $src ?>
          </a>
          </li>
          <?php } ?>
      </ul>

      <h3>Mac</h3>

      <p>On Mac, Jobber can be installed with Homebrew:</p>

      <pre>brew install jobber
sudo brew services start jobber</pre>
      
      <h3>Linux Packages</h3>
      <table class="table">
        <thead>
          <tr>
            <th>OS</th>
            <th>Platform</th>
            <th>File</th>
            <th>Size</th>
          </tr>
        </thead>
        <tfoot></tfoot>
        <tbody>
          <?php foreach ($release["assets"] as $os => $asset) { ?>
          <tr>
          <td><?= $os ?></td>
          <td><?= $asset["CPU"] ?></td>
          <td>
              <a href="<?= $asset["url"] ?>">
              <span class="fa fa-download" aria-hidden="true">&nbsp;</span><?= $asset["name"] ?>
              </a>
          </td>
          <td><?= $asset["size"] ?></td>
          </tr>
          <?php } ?>
        </tbody>
      </table>
      
      <h3>Docker Images</h3>
      
      <pre>docker pull jobber</pre>
      
      <p>These images contain Jobber running as an unprivileged user named &ldquo;jobberuser&rdquo;.
        The jobs are defined in the file <code>/home/jobberuser/.jobber</code>. By default, the
        only job is one that prints &ldquo;Jobber is running!&rdquo; every second. You should
        replace it with your own jobs.</p>

      <p>The images are based on Alpine. You can find a list of them
        <a href="https://hub.docker.com/_/jobber">here</a>.</p>
      
    </section>
  </div>
  <!-- main content -->
</body>

</html>