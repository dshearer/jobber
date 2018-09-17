<?php


function _errExit($msg)
{
    echo($msg);
    exit(1);
}

function _formatSize($bytes)
{
    $MB = 1 << 20;
    return sprintf("%.2f MB", $bytes/$MB);
}

function _reformatDate($date)
{
    date_default_timezone_set("UTC");
    $dateParts = date_parse($date);
    $unix = mktime(
        $dateParts["hour"],
        $dateParts["minute"],
        $dateParts["second"],
        $dateParts["month"],
        $dateParts["day"],
        $dateParts["year"]
    );
    if ($unix === FALSE) {
        _errExit("Failed to parse date\n");
    }
    date_default_timezone_set("America/Los_Angeles");
    return date("j M Y", $unix);
}

function _strEndsWith($haystack, $needle)
{
    $length = strlen($needle);
    if ($length == 0) {
        return true;
    }

    return (substr($haystack, -$length) === $needle);
}

function _findPlatform($asset)
{
    $PLATFORMS = [
        'el6.x86_64.rpm' => [
            'OS' => "RHEL/CentOS 6",
            'CPU' => "x86_64"
        ],
        'el7.x86_64.rpm' => [
            'OS' => "RHEL/CentOS 7",
            'CPU' => "x86_64"
        ],
        '.apk' => [
            'OS' => "Alpine Linux 3",
            'CPU' => "x86_64"
        ],
        '.deb' => [
            'OS' => "Debian 8+ / Ubuntu 14.10+",
            'CPU' => "x86_64"
        ]
    ];
    $filename = $asset["name"];
    foreach ($PLATFORMS as $suffix => $platform)
    {
        if (_strEndsWith($filename, $suffix)) {
            return $platform;
        }
    }

    _errExit("No platform for {$filename}\n");
}

function latestRelease()
{
    $INFO_PATH = "phplib/latest-release.json";

    $tmp = file_get_contents($INFO_PATH, TRUE);
    if ($tmp === FALSE) {
        _errExit("Failed to open {$INFO_PATH}\n");
    }
    $raw_info = json_decode($tmp, true);
    if ($raw_info === FALSE) {
        _errExit("Failed to parse {$INFO_PATH}\n");
    }
    if ($raw_info["prerelease"]) {
        _errExit("Release is prerelease!\n");
    }
    if ($raw_info["draft"]) {
        _errExit("Release is draft!\n");
    }

    // make final info structure
    $final_info = [
        "name" => $raw_info["name"],
        "date" => _reformatDate($raw_info["published_at"]),
        "rel_notes_url" => $raw_info["html_url"],
        "zipball_url" => $raw_info["zipball_url"],
        "tarball_url" => $raw_info["tarball_url"]
    ];

    // make assets
    $new_assets = array();
    foreach ($raw_info["assets"] as $asset)
    {
        $platform = _findPlatform($asset);
        $new_asset = [
            "CPU" => $platform["CPU"],
            "name" => $asset["name"],
            "size" => _formatSize($asset["size"]),
            "url" => $asset["browser_download_url"]
        ];
        $new_assets[$platform["OS"]] = $new_asset;
    }
    ksort($new_assets, SORT_NATURAL|SORT_FLAG_CASE);
    $final_info["assets"] = $new_assets;

    return $final_info;
}

?>
