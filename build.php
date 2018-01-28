<?
function errExit($msg)
{
    echo($msg);
    exit(1);
}

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
    $basePath = substr($path, 0, fileExtPos($path));
    $destPath = $basePath . "html";
    $desc = [
        1 => ["file", $destPath, "w"]
    ];
    $cmd = PHP_BINARY . " " . $path;
    $proc = proc_open($cmd, $desc, $pipes);
    if (!is_resource($proc)) {
        errExit("Failed to spawn PHP subproc\n");
    }
    $retval = proc_close($proc);
    if ($retval !== 0) {
        errExit("ERROR while making " . $destPath . "\n");
    }
    echo "Made: " . $destPath . "\n";
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
    $files = scandir($srcDir) or errExit("scandir failed\n");
    foreach ($files as $file)
    {
        $filePath = $srcDir . "/" . $file;
        if ($file == "." || $file == ".." || $file == "build.php" || 
            strpos($filePath, "phplib") !== FALSE) 
        {
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
    }
}

processSrc(".");
?>