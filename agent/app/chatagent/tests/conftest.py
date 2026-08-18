import sys
from pathlib import Path
from unittest.mock import MagicMock, patch

from botocore.exceptions import ClientError

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))


def _not_found_error() -> ClientError:
    return ClientError(
        error_response={"Error": {"Code": "ParameterNotFound", "Message": "not found"}},
        operation_name="GetParameter",
    )


# main.py（およびそこからimportされるmodel/load.py）はモジュールレベルでSSMに
# アクセスするため、テストファイルが main を初めてimportする前（=このconftest.py自体の
# import時点）でboto3.clientをモックしておく必要がある。フィクスチャにすると
# テスト収集（各テストファイルのimport）より後に実行されてしまい間に合わない。
_mock_ssm_client = MagicMock()
_mock_ssm_client.get_parameter.side_effect = _not_found_error()
with patch("boto3.client", return_value=_mock_ssm_client):
    import main  # noqa: F401,E402
