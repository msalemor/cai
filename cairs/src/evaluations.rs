use std::{collections::HashMap, fs};

use serde::{Deserialize, Serialize};

//use serde::{Deserialize, Serialize};
//use std::collections::HashMap;
#[derive(Deserialize, Serialize, Debug, Clone)]
pub struct Evaluation {
    pub name: String,
    pub description: String,
    #[serde(rename = "systemPrompt")]
    pub system_prompt: String,
}

pub fn load_evaluations() -> Result<HashMap<String, Evaluation>, Box<dyn std::error::Error>> {
    print!("Loading evaluations from src/evaluations.json... ");
    let content = fs::read_to_string("./evaluations.json")?;
    let evaluations: Vec<Evaluation> = serde_json::from_str(&content)?;

    let mut eval_map = HashMap::new();
    for eval in evaluations {
        eval_map.insert(eval.name.clone(), eval);
    }

    Ok(eval_map)
}
