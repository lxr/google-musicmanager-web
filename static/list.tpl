<!DOCTYPE html>
<html>
	<head>
		{{$purchasedOnly := printf "%s" .XXX_unrecognized}}
		<meta charset="UTF-8">
		<title>{{if not $purchasedOnly}}All tracks{{else}}Purchased and promotional{{end}} - Google Play Music Web Manager</title>
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
			<h1>{{if not $purchasedOnly}}All tracks{{else}}Purchased and promotional{{end}}</h1>
			<p>
				<a href="?">All tracks</a> •
				<a href="?purchasedOnly=true">Purchased and promotional</a>
			</p>
			<p class="stats">
				{{len .GetDownloadTrackInfo}} songs •
				last updated {{(time .GetUpdatedMin).Format "2006-01-02T15:04:05Z07:00"}}
				{{if .GetContinuationToken}}
				• <a rel="next"
					data-token="{{.GetContinuationToken}}"
					href="?pageToken={{.GetContinuationToken}}&purchasedOnly={{$purchasedOnly}}"
				>Next page</a>
				{{end}}
			</p>
			<form method="post" enctype="multipart/form-data">
				<label>Upload a song (MP3)</label>
				<input type="file" required accept="audio/mpeg,.mp3" name="track">
				<input type="submit">
			</form>
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
			{{range $i, $track := reverse .GetDownloadTrackInfo}}
				<tr id="{{$track.GetId}}" ondblclick="this.querySelector('[download]').click()">
					<td headers="index">{{incr $i}}</td>
					<td headers="title">{{$track.GetTitle}}</td>
					<td headers="artist">{{$track.GetArtist}}</td>
					<td headers="album">{{$track.GetAlbum}}</td>
					<td headers="dl"><a download href="{{$track.GetId}}">⬇</a></td>
				</tr>
			{{end}}
			</tbody>
		</table>
	</body>
</html>
