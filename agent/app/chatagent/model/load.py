import boto3
from botocore.exceptions import BotoCoreError, ClientError
from strands.models.bedrock import BedrockModel

DEFAULT_MODEL_ID = "global.anthropic.claude-sonnet-5"


def _load_model_id() -> str:
    try:
        ssm = boto3.client("ssm")
        return ssm.get_parameter(Name="/chatagent/AgentModelId")["Parameter"]["Value"]
    except (ClientError, BotoCoreError):
        return DEFAULT_MODEL_ID


MODEL_ID = _load_model_id()


def load_model() -> BedrockModel:
    """Get Bedrock model client using IAM credentials."""
    return BedrockModel(model_id=MODEL_ID)
