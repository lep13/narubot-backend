# NaruBot

NaruBot is a Webex bot that conducts a 10-question quiz based on the Naruto anime series. The bot identifies which Naruto character you most resemble by asking questions and calculating a score based on your responses. Besides the quiz, NaruBot also utilizes Google Vertex AI for responding to non-quiz-related messages, making it a versatile conversational bot.

![1](https://github.com/user-attachments/assets/433af53f-8eab-4e82-b1ce-18923d8dde0a)
![2](https://github.com/user-attachments/assets/45dea9f4-0c18-4b79-859a-530bdfb794c0)
![3](https://github.com/user-attachments/assets/4b2d2a74-f176-4b84-90d0-308af2cbd7ea)
![4](https://github.com/user-attachments/assets/0811d987-4daa-416c-846a-1b6228787066)
![5](https://github.com/user-attachments/assets/226889cd-bf8b-4454-a4f2-577260a92d0b)

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Project Structure](#project-structure)
4. [Environment Configuration](#environment-configuration)
5. [Usage](#usage)
6. [How It Works](#how-it-works)
7. [Credits](#credits)

## Prerequisites
Before using NaruBot, you’ll need:
- **Webex Account**: To create a bot and configure the necessary Webex settings.
- **Google Cloud Platform Account**: To access Vertex AI and obtain credentials for generative responses.
- **AWS Account**: To use AWS Secrets Manager for securely storing sensitive information.
- **MongoDB Atlas**: To store user quiz session data.

### Setting up the Webex Bot
1. Go to the [Webex Developer Portal](https://developer.webex.com/).
2. Sign in, navigate to **My Apps**, and create a new bot. You’ll receive a **Bot Email** and **Access Token** that you will use in your secrets configuration.
3. Set up a webhook for the bot to listen for new messages. Go to **Create a Webhook**, select `resource:messages`, and configure the URL to point to your backend's endpoint, e.g., `https://yourapp.com/webhook`.

### Google Cloud Platform Setup
1. Set up a Google Cloud Project and enable the **Vertex AI API**.
2. Under **IAM & Admin**, create a **Service Account** for your bot, then download the service account key in JSON format.
3. Go to **APIs & Services** -> **Credentials**, and create an **API Key** for Vertex AI.
4. Store these credentials in AWS Secrets Manager (as described in the configuration section).

## Installation
1. **Clone the Repository**
   ```bash
   git clone https://github.com/lep13/narubot-backend
   cd narubot-backend
   ```

2. **Install Dependencies**
   ```bash
   go mod tidy
   ```

3. **Run the Application**
   ```bash
   go run main.go
   ```

## Project Structure
```
NARUBOT-BACKEND
│── config/
│   └── config.go       # Loads configuration values from AWS Secrets Manager
│── controllers/
│   └── webhook_controller.go # Handles incoming webhook events from Webex
│── db/
│   └── db.go           # MongoDB connection setup
│── models/
│   ├── cards.go        # Models for Webex adaptive cards
│   ├── config.go       # Defines config structure for bot
│   ├── quiz_session.go # Quiz session model for tracking progress
│── router/
│   └── router.go       # API routes
│── services/
│   ├── quiz_service.go # Handles quiz logic
│   ├── vertexai_service.go # Integrates with Vertex AI for responses
│   ├── webhook_service.go # Sends messages and cards to Webex
│── character_description.json # Character descriptions for quiz results
│── quiz_questions.json # Quiz questions and possible answers
│── go.mod, go.sum      # Go module dependencies
│── main.go             # Entry point of the application
```

## Environment Configuration
The bot relies on AWS Secrets Manager to securely store and retrieve sensitive information. Set up a secret named `webex_bot` containing the following keys:
  
- **BOT_EMAIL**: The email address of your Webex bot.
- **WEBEX_ACCESS_TOKEN**: The access token for the Webex bot.
- **SERVICE_ACCOUNT_KEY**: Google Service Account Key in JSON format.
- **GOOGLE_PROJECT_ID**: Google Cloud project ID.
- **GOOGLE_MODEL_ID**: Vertex AI model ID for generating responses.
- **GOOGLE_REGION**: Region where Vertex AI is hosted (e.g., `us-central1`).
- **GENAI_ACCESS_TOKEN**: API key for Vertex AI.
- **MONGO_URI**: MongoDB connection string.

## Usage
1. **Start the Quiz**: Users can type `quiz` to receive instructions on starting the quiz. After typing `start`, the quiz begins, and the bot will guide them through each question.
2. **Quiz Responses**: For each question, users can respond with numbers (`1`, `2`, `3`, or `4`). They can also type `quit` to end the quiz.
3. **Non-Quiz Messages**: For any non-quiz-related messages, the bot generates a response using Google Vertex AI. This works for casual chats and general questions.

## How It Works
### Webhook Handling
1. The bot is connected to Webex through a webhook that listens for `messages` events.
2. When a user sends a message, Webex triggers a webhook to the backend with the message details.
3. The webhook controller checks if the message is quiz-related or a general message.
4. If the message is quiz-related, the bot processes quiz logic and returns the next question or result.
5. If the message is general, it is sent to Google Vertex AI for a response.

### Quiz Handling
1. The quiz is stored as a JSON file (`quiz_questions.json`).
2. User responses are tracked in a MongoDB session (`quiz_session.go`).
3. After all 10 questions, the bot calculates which Naruto character matches the user’s answers and sends the result as an adaptive card.

### Using Ngrok for Local Testing
If testing locally, Webex requires a public URL for the webhook.
1. Install ngrok:
   ```bash
   brew install ngrok  # For MacOS
   choco install ngrok # For Windows
   ```
2. Start ngrok:
   ```bash
   ngrok http 8080
   ```
3. Use the generated public URL for setting up your Webex webhook.

## Credits
This bot was developed using the following:
- **Webex API** for messaging and card interactions.
- **Google Vertex AI** for generating conversational responses.
- **AWS Secrets Manager** for secure storage of configuration secrets.
- **MongoDB** for storing user quiz session data.
