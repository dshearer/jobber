// Define the `phonecatApp` module
var app = angular.module('jobberSite', ['ngCookies']);

var gReleaseData = ['html_url', 'name', 'body', 'created_at'];
var gAssetData = ['browser_download_url', 'name'];
var gAppVer = 1;
var gCookieVar = 'jobberAppState';

app.controller('JobberSiteController',
  function($scope, $http, $cookies) {
	/*
	 * NOTE: GitHub throttles API requests to max 60/hour.
	 */
	
	// init cookie
	var cookie = $cookies.getObject(gCookieVar);
	if (cookie === undefined || cookie.appVer < gAppVer) {
		cookie = {appVer: gAppVer, latestRelease: null};
	}
	
	// init model
	$scope.latestRelease = cookie.latestRelease;

    // get latest release
	var httpConfig = {
		cache: false,
		headers: {'If-None-Match': null, 'If-Modified-Since': null}
	};
	if (cookie.latestReleaseEtag !== undefined) {
		httpConfig.headers['If-None-Match'] = cookie.latestReleaseEtag;
	}
	$http.get('https://api.github.com/repos/dshearer/jobber/releases/latest',
			  httpConfig)
		.then(function success(resp) {
				// save release info
				cookie.latestRelease = makeRelease(resp);
				
				// save etag
				var etag = resp.headers('ETag');
				if (etag === undefined || etag === null) {
					delete cookie.latestReleaseEtag;
				}
				else {
					cookie.latestReleaseEtag = etag;
				}
				
				// save cookie
				$cookies.putObject(gCookieVar, cookie);
				
				// set model
				$scope.latestRelease = cookie.latestRelease;
			},
			function error(resp) {
				if (resp.status != 304) {
					console.log("HTTP error: " + resp.statusText);
				}
			});
});

app.filter('bytes', function() {
	return function(bytes, precision) {
		if (isNaN(parseFloat(bytes)) || !isFinite(bytes)) return '-';
		if (typeof precision === 'undefined') precision = 1;
		var units = ['bytes', 'KB', 'MB', 'GB', 'TB', 'PB'],
			number = Math.floor(Math.log(bytes) / Math.log(1024));
		return (bytes / Math.pow(1024, Math.floor(number))).toFixed(precision) +  ' ' + units[number];
	}
});

function copyAttrs(to, from, attrs)
{
	for (var key in attrs)
	{
		to[attrs[key]] = from[attrs[key]];
	}
}

function makeRelease(githubResp) {
	// keep only interesting data
	var latestRelease = {};
	copyAttrs(latestRelease, githubResp.data, gReleaseData);
	
	// make source artifacts
	latestRelease.sourceArtifacts = [];
	latestRelease.sourceArtifacts.push({
		browser_download_url: githubResp.data.tarball_url,
		name: 				  githubResp.data.name + '.tar'
	});
	latestRelease.sourceArtifacts.push({
		browser_download_url: githubResp.data.zipball_url,
		name: 				  githubResp.data.name + '.zip'
	});
	
	// make binary artifacts
	latestRelease.binaryArtifacts = [];
	for (var key in githubResp.data.assets)
	{
		var asset = githubResp.data.assets[key];
		if (asset.state != 'uploaded') {
			continue;
		}
		latestRelease.binaryArtifacts.push(makeBinaryArtifact(asset));
	}
	
	return latestRelease;
}

function makeBinaryArtifact(asset)
{
	var artifact = {
		browser_download_url: asset.browser_download_url,
		name: asset.name,
		size: asset.size
	}
	var tmp = artifact.name.split('.');
	artifact.platform = tmp[tmp.length - 2];
	var suffix = tmp[tmp.length - 1];
	switch (suffix)
	{
	case 'rpm':
		artifact.os = 'RHEL/CentOS/Fedora';
		break;
		
	case 'apk':
		artifact.os = 'Alpine Linux';
		break;
	
	case 'deb':
		artifact.os = 'Debian/Ubuntu';
		break;
		
	default:
		artifact.os = 'other';
		break;
	}
	return artifact;
}
