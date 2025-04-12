# Mocker

*Because sometimes innovation is just rebadging open source tools with a new name. Plus, why buy the cow when you get the milk for free?*

## Overview

Mocker is a Docker CLI plugin that extends Docker with AI model functionality. It provides a familiar syntax (`docker model`) using existing open-source tools with no gimmicks, subscriptions, or lock-in.

What does Mocker do? Well, it:
- Seamlessly integrates with Docker as a proper CLI plugin
- Runs AI models with minimal fuss and maximum transparency
- Works across all platforms
- Doesn't pretend to be doing anything more complex than connecting services that already exist

> "The best innovation is just connecting existing technologies and getting out of the way." — Anonymous Developer (probably)

## Quick Start

### Installing Mocker

Choose the installation method for your operating system:

#### Linux

```bash
# Download the latest release for your architecture (amd64 or arm64)
curl -L https://github.com/richardkiene/mocker/releases/latest/download/docker-model-linux-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz | tar xz

# Create the plugins directory if it doesn't exist
mkdir -p ~/.docker/cli-plugins

# Move the binary to the Docker CLI plugins directory
mv docker-model-linux-* ~/.docker/cli-plugins/docker-model

# Make it executable
chmod +x ~/.docker/cli-plugins/docker-model
```

#### macOS

```bash
# Download the latest release for your architecture (amd64 for Intel Macs or arm64 for Apple Silicon)
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/arm64/arm64/')
curl -L https://github.com/richardkiene/mocker/releases/latest/download/docker-model-darwin-$ARCH.tar.gz | tar xz

# Create the plugins directory if it doesn't exist
mkdir -p ~/.docker/cli-plugins

# Move the binary to the Docker CLI plugins directory
mv docker-model-darwin-* ~/.docker/cli-plugins/docker-model

# Make it executable
chmod +x ~/.docker/cli-plugins/docker-model
```

#### Windows (PowerShell)

```powershell
# Create the plugins directory if it doesn't exist
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.docker\cli-plugins"

# Download the latest release
Invoke-WebRequest -Uri "https://github.com/richardkiene/mocker/releases/latest/download/docker-model-windows-amd64.zip" -OutFile "$env:TEMP\docker-model.zip"

# Extract the zip file
Expand-Archive -Path "$env:TEMP\docker-model.zip" -DestinationPath "$env:TEMP"

# Move the binary to the Docker CLI plugins directory
Move-Item -Path "$env:TEMP\docker-model-windows-amd64.exe" -Destination "$env:USERPROFILE\.docker\cli-plugins\docker-model.exe"
```

### Running Your First Model

```bash
# Pull a model
docker model pull gemma3:1b
   
# Start a chat
docker model run gemma3:1b
```

That's it! No subscriptions. No API keys. No BS.

## Commands

### Status

Check whether your model runner is active:

```console
$ docker model status
Mocker Model Runner is active
```

### Help

View all commands (that we definitely invented from scratch):

```console
$ docker model help
Usage:  docker model COMMAND

Commands:
  list        List models available locally
  pull        Download a model from Docker Hub
  rm          Remove a downloaded model
  run         Run a model interactively or with a prompt
  status      Check if the model runner is running
  version     Show the current version
```

### Version

Check the current version of Mocker and the underlying Ollama engine:

```console
$ docker model version
Mocker version: 1.0.0
Ollama version: ollama version is 0.6.5
```

### Pull a model

Pull a model to your local environment (where you own and control it):

```console
$ docker model pull qwen2.5:0.5b
Pulling model qwen2.5:0.5b (this is just Ollama in disguise, but don't tell anyone)...
pulling manifest 
pulling c5396e06af29... 100% ▕████████████████▏ 397 MB                         
pulling 66b9ea09bd5b... 100% ▕████████████████▏   68 B                         
pulling eb4402837c78... 100% ▕████████████████▏ 1.5 KB                         
pulling 832dd9e00a68... 100% ▕████████████████▏  11 KB                         
pulling 005f95c74751... 100% ▕████████████████▏  490 B                         
verifying sha256 digest 
writing manifest 
success 
Downloaded: 794.02 MB
Model qwen2.5:0.5b pulled successfully (just like some other tools do, but we're honest about it)
```

### List available models

List all models in your environment (no mysterious cloud processing here):

```console
$ docker model list
+MODEL       PARAMETERS  QUANTIZATION    ARCHITECTURE  MODEL ID      CREATED     SIZE
+qwen2.5:0.5b 397.00 M    Q4_0            llama         a8b0c5157701 seconds ago 397 MB
+gemma3:1b   815.00 M    Q4_K_M          gemma3        8648f39daa8f hours ago   815 MB
```

### Run a model

Run a model with a one-time prompt:

```console
$ docker model run gemma3:1b "Hi"
Running with prompt (Ollama is doing all the work, but we'll take credit)...
Hi there! How's your day going so far? 😊 

Is there anything you'd like to chat about or need help with?
```

Or start an interactive chat session:

```console
$ docker model run gemma3:1b
Interactive chat mode started. Type 'Ctrl+C' to exit.
(What you're about to use is just Ollama's interface with our name on it)
>>> How much wood could a woodchuck chuck if a woodchuck could chuck wood?
This is a classic riddle! The answer is:

A woodchuck would chuck as much wood as a woodchuck could chuck if a woodchuck could chuck wood. 

It's a pun! The question is designed to be nonsensical. 😊 

Let me know if you'd like to try another riddle!

>>> /bye
```

### Remove a model

Remove a downloaded model (with no lingering cloud copies):

```console
$ docker model rm qwen2.5:0.5b
Model qwen2.5:0.5b removed successfully (and we didn't charge you a subscription for it)
```

Verify the model has been removed:

```console
$ docker model list
+MODEL       PARAMETERS  QUANTIZATION    ARCHITECTURE  MODEL ID      CREATED     SIZE
+gemma3:1b   815.00 M    Q4_K_M          gemma3        8648f39daa8f hours ago   815 MB
```

## How it works

Mocker creates an Ollama container to run AI models. When you use model commands, it interacts with this container. 

Unlike certain other solutions, Mocker is completely transparent about what it's doing - it's simply connecting Docker with Ollama in a convenient way. Some companies might call this "AI innovation" and charge a subscription.

## API Integration

Want to integrate AI into your own applications? Since Mocker is just running Ollama in a container, you can access the Ollama API directly at `http://localhost:11434`.

### Example: Generate text with curl

```bash
curl -X POST http://localhost:11434/api/generate -d '{
  "model": "gemma3:1b",
  "prompt": "Explain Docker in simple terms"
}'
```

### Example: Chat with Python

```python
import requests

def chat_with_model(model, message, context=None):
    url = "http://localhost:11434/api/chat"
    data = {
        "model": model,
        "messages": [{"role": "user", "content": message}]
    }
    
    if context:
        data["context"] = context
        
    response = requests.post(url, json=data)
    return response.json()

# Example usage
response = chat_with_model("gemma3:1b", "What's the capital of France?")
print(response["message"]["content"])
```

### Example: Node.js Integration

```javascript
const axios = require('axios');

async function generateWithAI(model, prompt) {
  try {
    const response = await axios.post('http://localhost:11434/api/generate', {
      model: model,
      prompt: prompt
    });
    
    return response.data.response;
  } catch (error) {
    console.error('Error generating AI response:', error);
    return null;
  }
}

// Example usage
generateWithAI('gemma3:1b', 'Write a haiku about Docker')
  .then(result => console.log(result));
```

For complete API documentation, refer to the [Ollama API documentation](https://github.com/ollama/ollama/blob/main/docs/api.md).

## Build and Development

### Using the Makefile

The project includes a Makefile that makes it easy to build for different platforms:

```bash
# Clone the repository
git clone https://github.com/richardkiene/mocker.git
cd mocker

# Build and install for your current platform
make install

# Build for all platforms (creates builds in the dist directory)
make release

# Build for specific platforms
make build-linux   # For Linux (amd64, arm64)
make build-mac     # For macOS (amd64, arm64)
make build-windows # For Windows (amd64)

# Clean build artifacts
make clean
```

The `make release` command will create builds for all supported platforms and package them as compressed archives in the `dist` directory.

### Manual Installation from Source

1. Build the plugin:
   ```bash
   go build -o docker-model
   ```

2. Copy it to your Docker CLI plugins directory:
   - Linux/macOS: `~/.docker/cli-plugins/`
   - Windows: `%USERPROFILE%\.docker\cli-plugins\`

3. Make sure it's executable (Linux/macOS):
   ```bash
   chmod +x ~/.docker/cli-plugins/docker-model
   ```

## License

[MIT License](LICENSE) - Because we believe in open source, just like the tools we're wrapping.