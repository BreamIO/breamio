<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8" />
		<meta http-equiv="X-UA-Compatible" content="IE=edge" />
		<title></title>
		<meta name="description" content="" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/normalize/3.0.0/normalize.min.css" />
		<style>
			/* Main style for calibration application */

			main {
				position: relative;
			}

			ul {
				background-color: lightgrey;
				/*z-index: 1;*/
				position: absolute;
				top: 0;
				right: 0;
				bottom: 0;
				left: 0;
				margin: auto;
				padding: 1em 0 1em 2em;
				width: 12em;
				line-height: 1em;
				border-radius: 1em;
				height: {{ len .EyeTrackers }}em;
			}

			li {
				color: darkgrey;
			}

			.done {
				color: green;
			}

			.in-progress {
				color: orange;
			}

			a {
				color: black;
				text-decoration: none;
				cursor: default;
			}

			a::after {
				content: " " attr(href);
			}

			a:hover {
				background-color: grey;
			}

			circle {
				fill: lightsteelblue;
				stroke: black;
				stroke-width: 1px;
			}
		</style>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.1.0/jquery.min.js" async="async"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/bluebird/1.0.0/bluebird.min.js" async="async"></script>
		<script async="async">
			// Main script
			(function main(times) {
				var $circle,
						$ = window.jQuery,
						points = [
							[10, 10],
							[90, 10],
							[50, 50],
							[90, 90],
							[10, 90],
							[50, 50]
						]
				if(!$ || !Promise.is) {
					console.log("Missing", times)
					return setTimeout(main, 0, times + 1)
				}
				console.log("Loaded both")

				$circle = $("circle")

				function requestFullscreen(elem) {
					if (elem.requestFullscreen) {
					  elem.requestFullscreen();
					} else if (elem.msRequestFullscreen) {
					  elem.msRequestFullscreen();
					} else if (elem.mozRequestFullScreen) {
					  elem.mozRequestFullScreen();
					} else if (elem.webkitRequestFullscreen) {
					  elem.webkitRequestFullscreen();
					}
				}

				function relative(start, end) {
					return end
					start = +this.attr(start).slice(0,-1)
					console.log("relative", start, end)
					if (start > end) {
						return "-=" + (start - end)
					} else {
						return "+=" + (end - start)
					}
				}

				function move(cx, cy) {
					var end = {
						cx: relative.call(this, "cx", cx),
						cy: relative.call(this, "cy", cy)
					}
					console.log(cx, cy, end)
					return this.animate(end, {
						duration: 1000,
						step: function stepper(now, fx) {
							console.log("stepper", fx.prop, now)
							$(this).attr(fx.prop, ""+now+"%")
						}
					})
				}

				function pulse() {
					return this.animate({
						cy: 0.1
					}, {
						duration: 200,
						step: function stepper(now, fx) {
							$(this).attr(fx.prop, now)
						}
					}).animate({
						cy: 1
					}, {
						duration: 200,
						step: function stepper(now, fx) {
							$(this).attr(fx.prop, ""+now+"cm")
						}
					})
				}

				window.next = function next(index) {
					return move.apply($circle, points[index])
				}

				$("ul").on("click", "a", function clicked(event) {
					var $target = $(event.target),
							nr = $target.attr("href").slice(1)
							i = 0
					event.preventDefault()

					$("ul").hide()
					$circle.show()
					requestFullscreen($("body")[0])
					next(0)

					// points.reduce(function folder(circle, point) {
					// 	return move.apply(circle, point)
					// 	// return pulse.call(move.apply(circle, point))
					// }, $circle)
				})
			})(0)
		</script>
	</head>
	<body>
		<main>
			<svg version="1.1" xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" viewPort="0 0 100% 100%">
				<circle style="display: none;" cx="50%" cy="50%" r="1cm" />
			</svg>
			<ul>
				{{ range $index, $element := .EyeTrackers }}
				<li{{ with $element }} class="{{ . }}"{{ end }}>
					<a href="#{{ $index }}">Calibrate EyeTracker</a>
				</li>
				{{ end }}
			</ul>
		</main>
	</body>
</html>
