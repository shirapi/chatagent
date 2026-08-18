from unittest.mock import MagicMock, patch

from botocore.exceptions import ClientError, NoCredentialsError

from model.load import DEFAULT_MODEL_ID, _load_model_id


def _not_found_error() -> ClientError:
    return ClientError(
        error_response={"Error": {"Code": "ParameterNotFound", "Message": "not found"}},
        operation_name="GetParameter",
    )


class TestLoadModelID:
    def test_returns_ssm_value_when_available(self):
        mock_client = MagicMock()
        mock_client.get_parameter.return_value = {"Parameter": {"Value": "custom-model-id"}}

        with patch("boto3.client", return_value=mock_client):
            got = _load_model_id()

        assert got == "custom-model-id"

    def test_falls_back_to_default_on_client_error(self):
        mock_client = MagicMock()
        mock_client.get_parameter.side_effect = _not_found_error()

        with patch("boto3.client", return_value=mock_client):
            got = _load_model_id()

        assert got == DEFAULT_MODEL_ID

    def test_falls_back_to_default_when_credentials_are_missing(self):
        # NoCredentialsErrorはClientErrorのサブクラスではないため、
        # BotoCoreErrorも捕捉していないと未処理のまま伝播してしまう。
        mock_client = MagicMock()
        mock_client.get_parameter.side_effect = NoCredentialsError()

        with patch("boto3.client", return_value=mock_client):
            got = _load_model_id()

        assert got == DEFAULT_MODEL_ID
