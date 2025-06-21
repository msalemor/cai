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

// Evaluation result structures
#[derive(Deserialize, Serialize, Debug)]
pub struct EvaluationResult {
    pub score: i32,
    pub explanation: String,
}

#[derive(Serialize, Debug)]
pub struct FileEvaluation {
    pub file_path: String,
    pub evaluation_name: String,
    pub result: EvaluationResult,
    pub timestamp: String,
}

#[derive(Serialize, Debug)]
pub struct EvaluationSummary {
    pub total_files: usize,
    pub average_score: f64,
    pub evaluation_name: String,
    pub target_folder: String,
    pub timestamp: String,
}

#[derive(Serialize, Debug)]
pub struct EvaluationReport {
    pub summary: EvaluationSummary,
    pub results: Vec<FileEvaluation>,
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

// Parse Azure OpenAI response to extract evaluation result
pub fn parse_evaluation_response(response: &str) -> Result<EvaluationResult, String> {
    // Try to find JSON in the response
    let json_start = response.find('{');
    let json_end = response.rfind('}');

    if let (Some(start), Some(end)) = (json_start, json_end) {
        let json_str = &response[start..=end];
        match serde_json::from_str::<EvaluationResult>(json_str) {
            Ok(result) => {
                // Validate score is within expected range
                if result.score >= 1 && result.score <= 10 {
                    Ok(result)
                } else {
                    Err(format!(
                        "Score {} is outside valid range 1-10",
                        result.score
                    ))
                }
            }
            Err(e) => Err(format!("Failed to parse JSON: {}, JSON: {}", e, json_str)),
        }
    } else {
        Err("No JSON found in response".to_string())
    }
}

// Save results in JUnit XML format
pub fn save_junit_results(
    report: &EvaluationReport,
    filename: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut xml = String::new();
    xml.push_str("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
    xml.push_str(&format!(
        "<testsuites name=\"Code Evaluation - {}\" tests=\"{}\" failures=\"0\" errors=\"0\" time=\"0\">\n",
        report.summary.evaluation_name, report.summary.total_files
    ));
    xml.push_str(&format!(
        "  <testsuite name=\"{}\" tests=\"{}\" failures=\"0\" errors=\"0\" time=\"0\">\n",
        report.summary.evaluation_name, report.summary.total_files
    ));

    for result in &report.results {
        xml.push_str(&format!(
            "    <testcase classname=\"{}\" name=\"{}\" time=\"0\">\n",
            result.evaluation_name, result.file_path
        ));
        xml.push_str(&format!(
            "      <system-out>Score: {}/10\nExplanation: {}</system-out>\n",
            result.result.score, result.result.explanation
        ));
        xml.push_str("    </testcase>\n");
    }

    xml.push_str("  </testsuite>\n");
    xml.push_str("</testsuites>\n");

    std::fs::write(filename, xml)?;
    Ok(())
}

// Save results as JSON
pub fn save_json_results(
    report: &EvaluationReport,
    filename: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let json = serde_json::to_string_pretty(report)?;
    std::fs::write(filename, json)?;
    Ok(())
}
