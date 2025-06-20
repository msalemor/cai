use serde::{Deserialize, Serialize};

// Azure OpenAI API structures
#[derive(Serialize, Deserialize, Debug)]
struct ChatMessage {
    role: String,
    content: String,
}

#[derive(Serialize, Debug)]
struct ChatCompletionRequest {
    model: String, // This should be set to the deployment name
    messages: Vec<ChatMessage>,
    max_tokens: Option<i32>,
    temperature: Option<f32>,
}

#[derive(Deserialize, Debug)]
struct ChatCompletionChoice {
    message: ChatMessage,
    //index: i32,
    //finish_reason: Option<String>,
}

#[derive(Deserialize, Debug)]
struct ChatCompletionResponse {
    choices: Vec<ChatCompletionChoice>,
    //id: String,
    //object: String,
    //created: i64,
}

// Azure OpenAI chat completion function
pub async fn call_azure_openai(
    endpoint: &str,
    api_key: &str,
    model: &str,
    call_type: &str,
    system_prompt: &str,
    user_message: &str,
) -> Result<String, Box<dyn std::error::Error>> {
    let client = reqwest::Client::new();
    let url = format!("{}", endpoint);

    let messages = vec![
        ChatMessage {
            role: "system".to_string(),
            content: system_prompt.to_string(),
        },
        ChatMessage {
            role: "user".to_string(),
            content: user_message.to_string(),
        },
    ];

    let request_body = ChatCompletionRequest {
        model: model.to_string(),
        messages,
        max_tokens: Some(1000),
        temperature: Some(0.7),
    };

    let header_key;

    if call_type == "openai" {
        // For Azure OpenAI, we need to specify the deployment name in the model field
        header_key = "Bearer".to_string();
    } else {
        header_key = "api-key".to_string();
    }

    let response = client
        .post(&url)
        .header(&header_key, api_key)
        .header("Content-Type", "application/json")
        .json(&request_body)
        .send()
        .await?;

    if !response.status().is_success() {
        let error_text = response.text().await?;
        return Err(format!("Azure OpenAI API error: {}", error_text).into());
    }

    let completion_response: ChatCompletionResponse = response.json().await?;

    if let Some(choice) = completion_response.choices.first() {
        Ok(choice.message.content.clone())
    } else {
        Err("No response from Azure OpenAI".into())
    }
}
