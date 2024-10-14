
# NaruBot

NaruBot is a Webex bot that conducts a 10-question quiz based on the Naruto anime series. The bot identifies which Naruto character you most resemble by asking questions and calculating a score based on your responses. Besides the quiz, NaruBot also utilizes Google Vertex AI for responding to non-quiz-related messages, making it a versatile conversational bot.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Project Structure](#project-structure)
4. [Environment Configuration](#environment-configuration)
5. [Usage](#usage)
6. [Credits](#credits)

## Prerequisites
Before using NaruBot, you’ll need:
- **Webex Account**: To create a bot and configure the necessary Webex settings.
- **Google Cloud Platform Account**: To access Vertex AI and obtain credentials for generative responses.
- **AWS Account**: To use AWS Secrets Manager for securely storing sensitive information.
  
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
- **config/config.go**: Loads configuration values, especially those stored in AWS Secrets Manager.
- **controllers/webhook_controller.go**: Handles incoming webhook events from Webex, processes quiz and non-quiz-related messages, and sends responses.
- **db/db.go**: Manages the MongoDB connection and retrieves collections for user quiz sessions.
- **models/cards.go**: Contains models for Webex card structures, including helper functions for creating adaptive cards.
- **models/config.go**: Defines the configuration structure for the bot, integrating data from AWS Secrets Manager.
- **models/quiz_session.go**: Contains models for the quiz session, including structs for questions, options, and session tracking.
- **router/router.go**: Sets up the router and initializes routes for the Webex bot webhook.
- **services/quiz_service.go**: Manages the quiz logic, including starting a quiz, tracking answers, calculating results, and finalizing the quiz.
- **services/vertexai_service.go**: Communicates with Google Vertex AI to generate responses for non-quiz messages.
- **services/webhook_service.go**: Sends messages and adaptive cards to Webex.
- **character_description.json**: Contains character descriptions for the quiz result.
- **quiz_questions.json**: Holds the quiz questions and possible answers.

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

### Steps to Obtain Google AI Credentials
1. **Google Project ID**: Found in your Google Cloud Console under the project details.
2. **Vertex AI Model ID**: Create a model in Vertex AI, and note the model ID from the dashboard.
3. **Service Account Key**: Download the service account key from Google Cloud IAM, and store it as `SERVICE_ACCOUNT_KEY`.
4. **API Key for Vertex AI**: Obtain from Google Cloud Console under **APIs & Services** -> **Credentials**.

### Sample `config.json` for Local Testing
For local testing, you may skip AWS Secrets Manager and store your credentials in a `config.json` file in the root directory:
```json
{
  "BOT_EMAIL": "your_bot_email",
  "WEBEX_ACCESS_TOKEN": "your_access_token",
  "SERVICE_ACCOUNT_KEY": "your_service_account_key",
  "GOOGLE_PROJECT_ID": "your_google_project_id",
  "GOOGLE_MODEL_ID": "your_google_model_id",
  "GOOGLE_REGION": "your_google_region",
  "GENAI_ACCESS_TOKEN": "your_genai_access_token",
  "MONGO_URI": "your_mongo_uri"
}
```

## Usage
1. **Start the Quiz**: Users can type `quiz` to receive instructions on starting the quiz. After typing `start`, the quiz begins, and the bot will guide them through each question.
2. **Quiz Responses**: For each question, users can respond with numbers (`1`, `2`, `3`, or `4`). They can also type `quit` to end the quiz.
3. **Non-Quiz Messages**: For any non-quiz-related messages, the bot generates a response using Google Vertex AI. This works for casual chats and general questions.

## Credits
This bot was developed using the following:
- **Webex API** for messaging and card interactions.
- **Google Vertex AI** for generating conversational responses.
- **AWS Secrets Manager** for secure storage of configuration secrets.
- **MongoDB** for storing user quiz session data.
