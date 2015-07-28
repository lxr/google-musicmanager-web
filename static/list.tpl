<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{if not .PurchasedOnly}}All tracks{{else}}Purchased and promotional{{end}} - Google Play Music Web Manager</title>
		<script src="/static/sortable.js"></script>
		<link rel="icon" href="/static/icon.png">
		<link rel="stylesheet" href="/static/style.css">
		<script>
		window.onload = function () {
			Sortable.init()
			window.parent.postMessage(document.body.scrollHeight, "*")
		}
		</script>
	</head>
	<body id="list">
		<img class="cover" src="/static/icon.png" alt="">
		<div class="info">
			<h1>{{if not .PurchasedOnly}}All tracks{{else}}Purchased and promotional{{end}}</h1>
			<p>
				<a href="?">All tracks</a> •
				<a href="?purchasedOnly=true">Purchased and promotional</a>
			</p>
			<p class="stats">
				{{len .Items}} songs •
				Last modified {{(time .UpdatedMin).Format "2006-01-02T15:04:05Z07:00"}}
				{{if .PageToken}}
				• <a rel="next"
					data-token="{{.PageToken}}"
					href="?pageToken={{.PageToken}}&purchasedOnly={{.PurchasedOnly}}"
				>Next page</a>
				{{end}}
			</p>
		</div>
		<table id="response" data-sortable>
			<thead>
				<tr>
					<th id="index">#</th>
					<th id="title">Name</th>
					<th id="artist">Artist</th>
					<th id="album">Album</th>
					<th id="dl" data-sortable=false>DL</th>
				</tr>
			</thead>
			<tbody>
			{{range $i, $track := .Items}}
				<tr id="{{$track.Id}}" ondblclick="this.querySelector('[download]').click()">
					<td headers="index">{{incr $i}}</td>
					<td headers="title">{{$track.Title}}</td>
					<td headers="artist">{{$track.Artist}}</td>
					<td headers="album">{{$track.Album}}</td>
					<td headers="dl"><a download href="{{$track.Id}}">⬇</a></td>
				</tr>
			{{end}}
			</tbody>
		</table>
	</body>
</html>
