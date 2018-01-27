<?
function fileExtPos($path)
{
    $dotPos = strrpos($path, ".");
    if ($dotPos < 0) {
        return $dotPos;
    }
    else {
        return $dotPos + 1;
    }
}

function processPhpSrc($path)
{
    echo "PHP file: " . $path . "\n";
    $basePath = substr($path, 0, fileExtPos($path));
    $destPath = $basePath . "html";
    echo "Dest: " . $destPath . "\n";
    $desc = [
        1 => ["file", $destPath, "w"]
    ];
    $cmd = PHP_BINARY . " " . $path;
    $proc = proc_open($cmd, $desc, $pipes);
    if (!is_resource($proc)) {
        die("Failed to spawn PHP subproc\n");
    }
    proc_close($proc);
}

function processNonPhpSrc($path)
{
    // NOP
}

function isPhpFile($path)
{
    $extPos = fileExtPos($path);
    if ($extPos < 0) {
        return false;
    }
    else {
        return substr($path, $extPos) == "php";
    }
}

// process source files
function processSrc($srcDir)
{
    $files = scandir($srcDir) or die("scandir failed\n");
    foreach ($files as $file)
    {
        $filePath = $srcDir . "/" . $file;
        if ($file == "." || $file == ".." || $file == "build.php") {
            continue;
        }
        elseif (is_dir($filePath)) {
            processSrc($filePath);
        }
        elseif (is_file($filePath)) {
            if (isPhpFile($filePath)) {
                processPhpSrc($filePath);
            }
            else {
                processNonPhpSrc($filePath);
            }
        }
        else {
            echo("Ignoring " . $filePath);
        }
    }
}

processSrc(".");
?>