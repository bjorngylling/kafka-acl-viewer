package main

import "html/template"

type graphTemplateData struct {
	Nodes template.JS
	Edges template.JS
}

var graphTemplate = `<!doctype html>
<html>
<head>
<title>Kafka topic layout</title>

<script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/vis/4.21.0/vis.min.js"></script>
<link href="https://cdnjs.cloudflare.com/ajax/libs/vis/4.21.0/vis-network.min.css" rel="stylesheet" type="text/css" />

<style type="text/css">
#mynetwork {
width: 1600px;
height: 1080px;
}
</style>
</head>
<body>

<div id="mynetwork"></div>

<script type="text/javascript">
new vis.Network(
document.getElementById('mynetwork'),
{nodes: new vis.DataSet({{.Nodes}}), edges: new vis.DataSet({{.Edges}})},
{});
</script>


</body>
</html>`
