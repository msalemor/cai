pub mod evaluations;
pub mod openai;
pub mod util;

use clap::{Parser, Subcommand};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Command line tool with ls and evaluate subcommands
#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    /// List directory contents
    Ls,
    /// Evaluate with specified parameters
    Evaluate {
        /// Target folder path
        #[arg(long)]
        target_folder: String,

        /// Evaluation name
        #[arg(long)]
        evaluation_name: String,

        /// File to skip during evaluation
        #[arg(long)]
        skip_files: Option<String>,

        /// Files to include in evaluation
        #[arg(long)]
        include_files: Option<String>,

        /// JUnit file name for output
        #[arg(long)]
        junit_file_name: Option<String>,
    },
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    // Load evaluations at startup
    let evaluations = match evaluations::load_evaluations() {
        Ok(evals) => {
            println!("Loaded {} evaluations", evals.len());
            evals
        }
        Err(e) => {
            eprintln!("Error loading evaluations: {}", e);
            std::process::exit(1);
        }
    };

    match args.command {
        Commands::Ls => {
            println!("Executing ls command...");
            handle_ls(&evaluations);
        }
        Commands::Evaluate {
            target_folder,
            evaluation_name,
            skip_files: skip_file,
            include_files,
            junit_file_name,
        } => {
            println!("Executing evaluate command...");
            handle_evaluate(
                target_folder,
                evaluation_name,
                skip_file,
                include_files,
                junit_file_name,
                &evaluations,
            )
            .await;
        }
    }
}

fn handle_ls(evaluations: &HashMap<String, evaluations::Evaluation>) {
    println!("Listing available evaluations:");
    for (name, eval) in evaluations {
        println!("  {}: {}", name, eval.description);
    }
}

// Helper function to apply file filters
fn apply_file_filters(
    files: &[String],
    skip_pattern: Option<&str>,
    include_pattern: Option<&str>,
) -> Vec<String> {
    let mut filtered = files.to_vec();

    // Apply skip pattern filter
    if let Some(skip) = skip_pattern {
        filtered.retain(|file| !matches_pattern(file, skip));
    }

    // Apply include pattern filter
    if let Some(include) = include_pattern {
        let patterns: Vec<&str> = include.split(',').map(|s| s.trim()).collect();
        filtered.retain(|file| {
            patterns
                .iter()
                .any(|pattern| matches_pattern(file, pattern))
        });
    }

    filtered
}

// Simple pattern matching for file extensions and basic patterns
fn matches_pattern(file: &str, pattern: &str) -> bool {
    if pattern.starts_with("*.") {
        let ext = &pattern[2..];
        file.ends_with(&format!(".{}", ext))
    } else if pattern.contains('*') {
        // Basic wildcard support - convert to regex-like matching
        let pattern_parts: Vec<&str> = pattern.split('*').collect();
        if pattern_parts.len() == 2 {
            file.starts_with(pattern_parts[0]) && file.ends_with(pattern_parts[1])
        } else {
            file.contains(pattern)
        }
    } else {
        file.contains(pattern)
    }
}

// Evaluation result structures
#[derive(Deserialize, Serialize, Debug)]
struct EvaluationResult {
    score: i32,
    explanation: String,
}

#[derive(Serialize, Debug)]
struct FileEvaluation {
    file_path: String,
    evaluation_name: String,
    result: EvaluationResult,
    timestamp: String,
}

#[derive(Serialize, Debug)]
struct EvaluationReport {
    summary: EvaluationSummary,
    results: Vec<FileEvaluation>,
}

#[derive(Serialize, Debug)]
struct EvaluationSummary {
    total_files: usize,
    average_score: f64,
    evaluation_name: String,
    target_folder: String,
    timestamp: String,
}

async fn handle_evaluate(
    target_folder: String,
    evaluation_name: String,
    skip_file: Option<String>,
    include_files: Option<String>,
    junit_file_name: Option<String>,
    evaluations: &HashMap<String, evaluations::Evaluation>,
) {
    // Check if the evaluation exists
    let prompt;
    if let Some(evaluation) = evaluations.get(&evaluation_name) {
        prompt = evaluation.system_prompt.clone();
    } else {
        eprintln!("Error: Evaluation '{}' not found", evaluation_name);
        println!("Available evaluations:");
        for name in evaluations.keys() {
            println!("  {}", name);
        }
        return;
    }

    println!("🔍 Starting evaluation: {}", evaluation_name);
    println!("📁 Target folder: {}", target_folder);

    if let Some(skip) = &skip_file {
        println!("⏭️  Skip files pattern: {}", skip);
    }

    if let Some(include) = &include_files {
        println!("📋 Include files pattern: {}", include);
    }

    let source_code_files = util::list_source_files(&target_folder);

    // Apply filters if specified
    let filtered_files = apply_file_filters(
        &source_code_files,
        skip_file.as_deref(),
        include_files.as_deref(),
    );

    if filtered_files.is_empty() {
        println!(
            "❌ No source files found matching criteria in '{}'",
            target_folder
        );
        return;
    }

    println!(
        "📄 Found {} source files to evaluate:",
        filtered_files.len()
    );
    for file in &filtered_files {
        println!("  • {}", file);
    }

    // Azure OpenAI configuration
    let azure_endpoint = std::env::var("AZURE_OPENAI_ENDPOINT").unwrap_or_else(|_| {
        println!("⚠️  Warning: AZURE_OPENAI_ENDPOINT environment variable not set");
        "https://your-resource.openai.azure.com".to_string()
    });

    let azure_api_key = std::env::var("AZURE_OPENAI_API_KEY").unwrap_or_else(|_| {
        println!("⚠️  Warning: AZURE_OPENAI_API_KEY environment variable not set");
        "your-api-key".to_string()
    });

    let deployment_name = std::env::var("AZURE_OPENAI_DEPLOYMENT_NAME").unwrap_or_else(|_| {
        println!("⚠️  Warning: AZURE_OPENAI_DEPLOYMENT_NAME environment variable not set");
        "gpt-4".to_string()
    });

    let mut evaluations_results = Vec::new();
    let timestamp = chrono::Utc::now().to_rfc3339();

    println!("\n🚀 Beginning AI evaluation...\n");

    // Process each source file with Azure OpenAI
    for (index, file_path) in filtered_files.iter().enumerate() {
        println!(
            "📝 Evaluating ({}/{}) {}",
            index + 1,
            filtered_files.len(),
            file_path
        );

        if let Ok(file_content) = std::fs::read_to_string(file_path) {
            let user_message = format!(
                "Please evaluate this source code file '{}':\n\n```\n{}\n```",
                file_path, file_content
            );

            match openai::call_azure_openai(
                &azure_endpoint,
                &azure_api_key,
                &deployment_name,
                &prompt,
                &user_message,
            )
            .await
            {
                Ok(response) => match parse_evaluation_response(&response) {
                    Ok(result) => {
                        println!("✅ Score: {}/10 - {}", result.score, result.explanation);

                        let file_eval = FileEvaluation {
                            file_path: file_path.clone(),
                            evaluation_name: evaluation_name.clone(),
                            result,
                            timestamp: timestamp.clone(),
                        };
                        evaluations_results.push(file_eval);
                    }
                    Err(e) => {
                        println!("❌ Failed to parse response: {}", e);
                        println!("   Raw response: {}", response);
                    }
                },
                Err(e) => {
                    println!("❌ Error calling Azure OpenAI: {}", e);
                }
            }
        } else {
            println!("❌ Error reading file: {}", file_path);
        }

        // Add a small delay between requests to be respectful to the API
        if index < filtered_files.len() - 1 {
            tokio::time::sleep(tokio::time::Duration::from_millis(500)).await;
        }
    }

    // Generate summary
    if !evaluations_results.is_empty() {
        let average_score = evaluations_results
            .iter()
            .map(|r| r.result.score as f64)
            .sum::<f64>()
            / evaluations_results.len() as f64;

        let summary = EvaluationSummary {
            total_files: evaluations_results.len(),
            average_score,
            evaluation_name: evaluation_name.clone(),
            target_folder: target_folder.clone(),
            timestamp: timestamp.clone(),
        };

        let report = EvaluationReport {
            summary,
            results: evaluations_results,
        };

        println!("\n📊 Evaluation Summary:");
        println!("   Total files evaluated: {}", report.summary.total_files);
        println!("   Average score: {:.1}/10", report.summary.average_score);

        // Save results if JUnit file specified
        if let Some(junit_file) = junit_file_name {
            if let Err(e) = save_junit_results(&report, &junit_file) {
                println!("❌ Failed to save JUnit results: {}", e);
            } else {
                println!("💾 Results saved to: {}", junit_file);
            }
        }

        // Also save as JSON report
        let json_file = format!(
            "evaluation_report_{}.json",
            chrono::Utc::now().format("%Y%m%d_%H%M%S")
        );
        if let Err(e) = save_json_results(&report, &json_file) {
            println!("❌ Failed to save JSON report: {}", e);
        } else {
            println!("💾 Detailed report saved to: {}", json_file);
        }
    } else {
        println!("\n❌ No successful evaluations completed.");
    }
}

// Parse Azure OpenAI response to extract evaluation result
fn parse_evaluation_response(response: &str) -> Result<EvaluationResult, String> {
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
fn save_junit_results(
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
fn save_json_results(
    report: &EvaluationReport,
    filename: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let json = serde_json::to_string_pretty(report)?;
    std::fs::write(filename, json)?;
    Ok(())
}
