<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8" />
		<meta http-equiv="X-UA-Compatible" content="IE=edge" />
		<title></title>
		<meta name="description" content="" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<link rel="stylesheet" href="//cdnjs.cloudflare.com/ajax/libs/normalize/3.0.0/normalize.min.css" />
		<style>
			/* Main style for calibration application */

			li {
				color: darkgrey;
			}

			li.done {
				color: green;
			}

			li.in-progress {
				color: orange;
			}

			li > a {
				color: black;
				text-decoration: none;
				cursor: default;
			}

			li > a::after {
				content: " " attr(href);
			}

			circle {
				fill: lightsteelblue;
				stroke: black;
				stroke-width: 1px;
			}
		</style>
		<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.1.0/jquery.js" async="async"></script>
		<script src="//cdnjs.cloudflare.com/ajax/libs/bluebird/1.0.0/bluebird.js" async="async"></script>
		<script defer="defer">
			// Main script
			console.log(jQuery, $)
			console.log(Promise)
		</script>
	</head>
	<body>
		<main>
			<ul>
				<li class="done">
					<a href="#1">Calibrate EyeTracker</a>
				</li>
				<li>
					<a href="#2">Calibrate EyeTracker</a>
				</li>
				<li class="in-progress">
					<a href="#3">Calibrate EyeTracker</a>
				</li>
			</ul>
			<svg version="1.1" xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" viewPort="0 0 100% 100%">
				<circle cx="50%" cy="50%" r="1cm" />
			</svg>
		</main>
	</body>
</html>
