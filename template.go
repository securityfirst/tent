package main

const tplLogin = `<html>
	<p>Not connected<p>
	<ul>
		<li><a href="%s">Login</a></li>
	</ul>
</html>`

const tplProfile = `<html>
	<p>Connected as <b>%s</b> &lt;%s&gt;</p>
	<ul>
		<li>Sample file <a href="%s">Edit</a> or <a href="%s">View</a> </li>
		<li><a href="%s">Logout</a></li>
	</ul>
</html>`

const tplEdit = `<html><head>
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/font-awesome/latest/css/font-awesome.min.css">
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/simplemde/latest/simplemde.min.css">
	<script src="https://cdn.jsdelivr.net/simplemde/latest/simplemde.min.js"></script>
	<script> document.addEventListener("DOMContentLoaded", function(){new SimpleMDE({element: document.getElementById("edit"),spellChecker: false});})</script>
</head><body>
	<p>Connected as <b>%s</b> &lt;%s&gt;</p>
	<p>Editing <b>%s</b></p>
	<h3>Previous Content</h3><pre>%s</pre>
	<h3>Edited Content</h3>
	<form method="POST" action="%s">
		<p><input type="submit" value="Save"/></p>
		<textarea id="edit" name="contents">%s</textarea><br/>
		<input type="hidden" name="hash" value="%s"/>
	</form>
</body></html>`
