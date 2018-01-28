<?php require("phplib/content-funcs.php"); ?>

<!DOCTYPE html>
<html lang="en">

<head>
<?php require("phplib/partials/head.html"); ?>

<title>Three Years of Go</title>
</head>

<body>

  <!-- NAV BAR -->
  <?php makeSubpageNavbar("blog"); ?>

  <!-- MAIN CONTENT -->
  <article class="container">
      <header>
        <h1>Three Years of Go</h1>
        <p>
          <small>C. Dylan Shearer | 14 Jan 2018</small>
        </p>
      </header>
      
    <p>I started Jobber more than three years ago. It was this first thing I
      had written in Go, and Jobber&rsquo;s need for threads made it a great
      project for learning this language.</p>

    <h2>I Like Go</h2>

    <p>Go is a simple, strongly-typed, memory-safe, compiled language that
      is good for systems programming. These qualities were in fact the main
      goals of the language, and so Go&rsquo;s designers have largely succeeded.</p>

    <p>Importantly, Go&rsquo;s concurrency constructs work very well.
      Passing messages instead of messing with locks is a great idea.</p>

    <p>Let me point out two particularly awesome traits: Go programs are
      compiled to native binaries, and Go libraries must be linked statically.
      Once you compile your code, your program and all its libraries are in
      exactly one file, and you can run that thing on any machine with the right
      kernel and CPU.</p>

    <p>This becomes a huge advantage over languages like Python, Ruby, and
      Javascript as soon as you start thinking "Gee, it would be nice to use
      this library in my little program." With Python et al., you have to worry
      not only about having the right version of the interpreter but also about
      having the right versions of all the libraries you want to use &mdash; and
      about them being in the right place!</p>

    <h2>What I Don&rsquo;t Like</h2>

    <h3>The Build Process</h3>

    <p>
      <code>GOPATH</code>
      , <a href="https://golang.org/doc/code.html#Organization">&ldquo;workspaces&rdquo;</a>,
      <code>go get</code>
      : these things do not work and need to die.
    </p>

    <p>I know how to use git; I do not need go to check crap out for me.</p>

    <p>
      It <em>should</em> be possible to just clone a repo, cd into the
      project&rsquo;s directory, and do
      <code>make</code>
      . Not with Go. Instead, you have to first <a
        href="/jobber/doc/v1.3/#deployment">futz around with making a
        workspace</a>, then change
      <code>GOPATH</code>
      , and then do
      <code>make</code>
      or
      <code>go build</code>
      or whatever.
    </p>

    <p>Go presumes that you have one workspace and don't mind checking out
      all sorts of crap into it, whereas people usually (or they should) prefer
      having different projects and their dependencies in different places.</p>

    <p>
      (Jobber needs to compile a Yacc grammar, and you may be interested in <a
        href="https://github.com/dshearer/jobber/blob/maint-1.3/mk/buildtools.mk">
        how I ensure goyacc is there during a build</a>.)
    </p>

    <p>
      Thank God they added support for the
      <code>vendor</code>
      dir.
    </p>

    <h3>
      Nitpick: Pointers and No
      <code>const</code>
    </h3>

    <p>I don&rsquo;t get why pointers were added to the language.
      Java&rsquo;s approach (where all object variables are really pointer
      variables) seems like it would work.</p>

    <p>
      And Go&rsquo;s equivalent of
      <code>void*</code>
      &mdash;
      <code>interface{}</code>
      &mdash; still doesn&rsquo;t make sense to me.
    </p>

    <p>
      I hope they add
      <code>const</code>
      , in the style of C++. If nothing else, it&rsquo;s a great way to document
      the side-effects of functions and methods.
    </p>
  </article>

  <!-- FOOTER  -->
  <footer class="small">
    <p>Copyright &#0169; 2018 C.&nbsp;Dylan Shearer</p>
  </footer>
</body>

</html>