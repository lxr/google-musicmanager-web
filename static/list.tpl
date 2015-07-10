<!DOCTYPE html>
<html id="list">
	<head>
		{{define "canon"}}{{if .}}canonical{{else}}alternate{{end}}{{end}}
		<meta charset="UTF-8">
		<title>{{if not .XXX_unrecognized}}All tracks{{else}}Purchased and promotional{{end}} - Google Play Music Web Manager</title>
		<script src="/static/sortable.js"></script>
		<link rel="shortcut icon" href="/static/icon.png">
		<link rel="stylesheet" href="/static/style.css">
		<link title="All tracks" href="?" rel="{{template "canon" not .XXX_unrecognized}}">
		<link title="Purchased and promotional" href="?purchasedOnly=true" rel="{{template "canon" .XXX_unrecognized}}">
		<meta name="songs" content="{{len .GetDownloadTrackInfo}}">
		{{if .GetContinuationToken}}
		<link title="Next page" rel="next"
			data-token="{{.GetContinuationToken}}"
			href="?pageToken={{.GetContinuationToken}}&purchasedOnly={{printf "%s" .XXX_unrecognized}}"
		>
		{{end}}
	</head>
	<body>
		<form method="post" enctype="multipart/form-data">
			<label>Upload a song (MP3)</label>
			<input type="file" required accept="audio/mpeg,.mp3" name="track">
			<input type="submit">
		</form>
		<table id="response" data-sortable>
			<thead>
				<tr>
					<th id="title">Name</th>
					<th id="artist">Artist</th>
					<th id="album">Album</th>
					<th id="dl" data-sortable=false>DL</th>
				</tr>
			</thead>
			<tbody>
			{{range reverse .GetDownloadTrackInfo}}
				<tr ondblclick="this.querySelector('[download]').click()">
					<td headers="title">{{.GetTitle}}</td>
					<td headers="artist">{{.GetArtist}}</td>
					<td headers="album">{{.GetAlbum}}</td>
					<td headers="dl"><a download href="{{.GetId}}">⬇</a></td>
				</tr>
			{{end}}
			</tbody>
		</table>
	</body>
</html>
