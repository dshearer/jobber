// Define the `phonecatApp` module
var app = angular.module('jobberSite', ['ngCookies']);

var gReleaseData = ['html_url', 'name', 'body', 'created_at'];
var gAssetData = ['browser_download_url', 'name'];

function copyAttrs(to, from, attrs)
{
	for (var key in attrs)
	{
		to[attrs[key]] = from[attrs[key]];
	}
}

app.controller('JobberSiteController',
  function($scope, $http, $cookies) {
	/*
	 * NOTE: GitHub throttles API requests to max 60/hour.
	 */
	
	// look for cached release info
	$scope.latestRelease = $cookies.getObject('latestRelease');
	if ($scope.latestRelease == undefined) {
	    // get releases
		$http.get('https://api.github.com/repos/dshearer/jobber/releases/latest')
			.then(function(resp) {
				// keep only interesting data
				$scope.latestRelease = {};
				copyAttrs($scope.latestRelease, resp.data, gReleaseData);
				
				// make source artifacts
				$scope.latestRelease.sourceArtifacts = [];
				$scope.latestRelease.sourceArtifacts.push({
					browser_download_url: resp.data.tarball_url,
					name: resp.data.name + '.tar'
				});
				$scope.latestRelease.sourceArtifacts.push({
					browser_download_url: resp.data.zipball_url,
					name: resp.data.name + '.zip'
				});
				
				// make binary artifacts
				$scope.latestRelease.binaryArtifacts = [];
				for (var key in resp.data.assets)
				{
					var asset = resp.data.assets[key];
					if (asset.state != 'uploaded') {
						continue;
					}
					var artifact = makeBinaryArtifact(asset);
					$scope.latestRelease.binaryArtifacts.push(artifact);
				}
				
				// save release info
				$cookies.putObject('latestRelease', $scope.latestRelease);
			});
	}
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
		
	default:
		artifact.os = 'other';
		break;
	}
	return artifact;
}
