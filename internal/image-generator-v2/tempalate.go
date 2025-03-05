package imagegeneratorv2

const strickerHtmlTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <style>
        html, body {
            width: 512px; height: 512px;
            margin: 0; padding: 0;
            background-color: #e0ddd5;
            font-family: sans-serif;
            display: flex;
            justify-content: center; align-items: center;
        }
        .chat {
            width: 480px; height: 480px;
            display: flex; flex-direction: column;
            justify-content: center;
        }
        .message {
            border-radius: 8px; padding: 8px 12px;
            width: fit-content; max-width: 85%;
            word-wrap: break-word; margin-bottom: 10px;
        }
        .incoming {
            background-color: #ffffff;
            align-self: flex-start;
            border-top-left-radius: 0;
        }
        .outgoing {
            background-color: #dcf8c6;
            align-self: flex-end;
            border-top-right-radius: 0;
        }
        .nickname {
            font-weight: bold;
            margin-bottom: 4px;
            color: #4f5b66;
            font-size: {{.FontSize}}px;
        }
        .text {
            font-size: {{.FontSize}}px;
            line-height: 1.2;
        }
    </style>
</head>
<body>
<div class="chat">
{{range .Messages}}
    <div class="message incoming">
        <div class="nickname">{{.UserNickname}}</div>
        <div class="text">{{.Text}}</div>
    </div>
{{end}}
</div>
</body>
</html>
`
