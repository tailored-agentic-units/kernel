# Azure AI Foundry Services and Models

To deploy a model for usage, you must do the following:

1. Create an [Azure AI Foundry account](https://learn.microsoft.com/en-us/cli/azure/cognitiveservices/account?view=azure-cli-latest&preserve-view=true#az-cognitiveservices-account-create), being sure to specify:
  - **kind** - the kind of account, which will determine the types of models that can be deployed to the account.
  - **location** - the Azure region to create the account in.
  - **name** - the name you want to use to reference the account you are creating.
  - **resource group** - the resource group to create the account in.
  - **sku** - the SKU you want to use for the account.

2. Create a [model deployment](https://learn.microsoft.com/en-us/cli/azure/cognitiveservices/account/deployment?view=azure-cli-latest&preserve-view=true#az-cognitiveservices-account-deployment-create), being sure to specify:
  - **model format** - the format of the model.
  - **model name** - the name of the model.
  - **model version** - the version of the model.
  - **name** - the cognitive services account name.
  - **resource group** - the resource group to create the deployment in.

> [!NOTE]
> Thanks to strange Microsoft naming conventions, there's a bit of naming weirdness for their AI platform. Cognitive Services has essentially been rebranded Azure OpenAI and the platform that hosts it is called Azure AI Foundry. To simplify my script naming, I've opted to use `foundry` in the script names instead of `cognitive-services` or `open-ai`.

> [!IMPORTANT]
> **Custom Subdomain Configuration**:
> - The infrastructure scripts now create a custom subdomain by default (e.g., `https://your-domain.openai.azure.com`)
> - This custom subdomain endpoint works with both API key and bearer token authentication
> - Bearer token authentication specifically requires custom subdomain endpoints, while API key auth can use either regional or custom subdomain endpoints

## Infrastructure Scripts

### [`create.sh`](./infrastructure/create.sh)

Creates a complete Azure infrastructure setup for AI services, including resource group, cognitive services account, and model deployment.

Parameter | Description | Default Value
----------|-------------|--------------
`--location` | Azure region for resources | `eastus`
`--resource-group` | Resource group name | `ATKResourceGroup`
`--cognitive-service` | Cognitive services account name | `ATKCognitiveService`
`--cognitive-services-sku` | SKU for cognitive services | `S0`
`--cognitive-service-kind` | Type of cognitive service | `OpenAI`
`--cognitive-service-domain` | Custom subdomain for the service | `go-agents-platform`
`--cognitive-service-role` | Role to grant for bearer token access | `Cognitive Services OpenAI User`
`--model-deployment` | Model deployment name | `o3-mini`
`--model-name` | AI model name | `o3-mini`
`--model-version` | Model version | `2025-01-31`
`--model-format` | Model format | `OpenAI`
`--model-sku` | Model SKU | `GlobalStandard`
`--model-sku-capacity` | Model capacity | `10`

### [`destroy.sh`](./infrastructure/destroy.sh)

Destroys Azure infrastructure by deleting the resource group and purging the cognitive services account.

Parameter | Description | Default Value
----------|-------------|--------------
`--resource-group` | Resource group to delete | `ATKResourceGroup`
`--location` | Azure region | `eastus`
`--cognitive-service` | Cognitive service name to purge | `ATKCognitiveService`

## Utility Scripts

### [`get-foundry-key.sh`](./utilities/get-foundry-key.sh)

Retrieves the API key for an Azure Cognitive Services account.

Parameter | Description | Default Value
----------|-------------|--------------
`--resource-group` | Resource group containing the service | `ATKResourceGroup`
`--cognitive-service` | Cognitive service account name | `ATKCognitiveService`

### [`get-foundry-endpoint.sh`](./utilities/get-foundry-endpoint.sh)

Retrieves the endpoint URL for an Azure Cognitive Services account.

Parameter | Description | Default Value
----------|-------------|--------------
`--resource-group` | Resource group containing the service | `ATKResourceGroup`
`--cognitive-service` | Cognitive service account name | `ATKCognitiveService`

### [`get-foundry-token.sh`](./utilities/get-foundry-token.sh)

Retrieves a bearer token for Azure Cognitive Services using Azure Entra ID authentication.

**Usage**:
```bash
# Get token and store in variable
AZURE_TOKEN=$(. scripts/azure/utilities/get-foundry-token.sh)

# Use token with prompt-agent
go run tools/prompt-agent/main.go \
  -config tools/prompt-agent/config.azure-entra.json \
  -token $AZURE_TOKEN \
  -prompt "Your prompt here"
```

**Note**: This requires being logged in to Azure CLI (`az login`) and having appropriate permissions on the Cognitive Services resource.

## Component Scripts

### [`resource-group.sh`](./components/resource-group.sh)

Creates an Azure resource group if it doesn't already exist.

Parameter | Description | Default Value
----------|-------------|--------------
`--resource-group` | Resource group name | `ATKResourceGroup`
`--location` | Azure region | `eastus`

### [`cognitive-services-account.sh`](./components/cognitive-services-account.sh)

Creates an Azure Cognitive Services account if it doesn't already exist.

Parameter | Description | Default Value
----------|-------------|--------------
`--kind` | Type of cognitive service |
`--location` | Azure region | `eastus`
`--name` | Service account name |
`--resource-group` | Resource group name |
`--sku` | Service SKU |
`--domain` | Custom subdomain for the service |

### [`cognitive-services-deployment.sh`](./components/cognitive-services-deployment.sh)

Creates a model deployment within an existing Azure Cognitive Services account.

Parameter | Description
----------|------------
`--model-format` | Format of the model 
`--model-name` | Name of the AI model 
`--model-version` | Version of the model 
`--name` | Cognitive services account name 
`--resource-group` | Resource group name 
`--deployment-name` | Name for the deployment 
`--sku` | SKU for the deployment 
`--sku-capacity` | Capacity for the SKU 

### [`cognitive-services-grant-permissions.sh`](./components/cognitive-services-grant-permissions.sh)

Grants the specified role to the current signed-in user on a Cognitive Services account. This is required for bearer token authentication.

Parameter | Description | Default Value
----------|-------------|--------------
`--name` | Cognitive services account name |
`--role` | Azure role to grant | `Cognitive Services OpenAI User`
`--resource-group` | Resource group name |

**Available Roles**:
- `Cognitive Services OpenAI User` - Basic access for using deployed models
- `Cognitive Services OpenAI Contributor` - Can manage deployments and models
- `Cognitive Services Contributor` - Full management access

### [`purge-cognitive-services-account.sh`](./components/purge-cognitive-services-account.sh)

Deletes and purges an Azure Cognitive Services account to ensure complete cleanup.

Parameter | Description | Default Value
----------|-------------|--------------
`--location` | Azure region | `eastus`
`--name` | Service account name to purge |
`--resource-group` | Resource group name |

## Query Scripts

I've created some scripts to make planning for this process a lot easier.

Note that all commands allow you to specify an `--output` argument that supports the following formats:

Format | Description
-------|------------
`json` | JSON string.
`jsonc` | Colorized json.
`table` | ASCII table with keys as column headings. This is the default in these scripts.
`tsv` | Tab-separated values, with no keys.
`yaml` | YAML, a human-readable alternative to JSON.
`yamlc` | Colorized YAML.
`none` | No output other than errors and warnings.

### [`locations.sh`](./queries/locations.sh)

Outputs the locations available within your current tenant.

`--output` is the only supported argument.

Example:

```sh
. scripts/azure/queries/locations.sh
```

<details>

  <summary>Output</summary>

  ```sh
  DisplayName               Name                 RegionalDisplayName
  ------------------------  -------------------  -------------------------------------
  East US                   eastus               (US) East US
  West US 2                 westus2              (US) West US 2
  Australia East            australiaeast        (Asia Pacific) Australia East
  Southeast Asia            southeastasia        (Asia Pacific) Southeast Asia
  North Europe              northeurope          (Europe) North Europe
  Sweden Central            swedencentral        (Europe) Sweden Central
  UK South                  uksouth              (Europe) UK South
  West Europe               westeurope           (Europe) West Europe
  Central US                centralus            (US) Central US
  South Africa North        southafricanorth     (Africa) South Africa North
  Central India             centralindia         (Asia Pacific) Central India
  East Asia                 eastasia             (Asia Pacific) East Asia
  Indonesia Central         indonesiacentral     (Asia Pacific) Indonesia Central
  Japan East                japaneast            (Asia Pacific) Japan East
  Japan West                japanwest            (Asia Pacific) Japan West
  Korea Central             koreacentral         (Asia Pacific) Korea Central
  Malaysia West             malaysiawest         (Asia Pacific) Malaysia West
  New Zealand North         newzealandnorth      (Asia Pacific) New Zealand North
  Canada Central            canadacentral        (Canada) Canada Central
  Austria East              austriaeast          (Europe) Austria East
  France Central            francecentral        (Europe) France Central
  Germany West Central      germanywestcentral   (Europe) Germany West Central
  Italy North               italynorth           (Europe) Italy North
  Norway East               norwayeast           (Europe) Norway East
  Poland Central            polandcentral        (Europe) Poland Central
  Spain Central             spaincentral         (Europe) Spain Central
  Switzerland North         switzerlandnorth     (Europe) Switzerland North
  Mexico Central            mexicocentral        (Mexico) Mexico Central
  UAE North                 uaenorth             (Middle East) UAE North
  Brazil South              brazilsouth          (South America) Brazil South
  Chile Central             chilecentral         (South America) Chile Central
  East US 2 EUAP            eastus2euap          (US) East US 2 EUAP
  Israel Central            israelcentral        (Middle East) Israel Central
  Qatar Central             qatarcentral         (Middle East) Qatar Central
  Central US (Stage)        centralusstage       (US) Central US (Stage)
  East US (Stage)           eastusstage          (US) East US (Stage)
  East US 2 (Stage)         eastus2stage         (US) East US 2 (Stage)
  North Central US (Stage)  northcentralusstage  (US) North Central US (Stage)
  South Central US (Stage)  southcentralusstage  (US) South Central US (Stage)
  West US (Stage)           westusstage          (US) West US (Stage)
  West US 2 (Stage)         westus2stage         (US) West US 2 (Stage)
  Asia                      asia                 Asia
  Asia Pacific              asiapacific          Asia Pacific
  Australia                 australia            Australia
  Brazil                    brazil               Brazil
  Canada                    canada               Canada
  Europe                    europe               Europe
  France                    france               France
  Germany                   germany              Germany
  Global                    global               Global
  India                     india                India
  Indonesia                 indonesia            Indonesia
  Israel                    israel               Israel
  Italy                     italy                Italy
  Japan                     japan                Japan
  Korea                     korea                Korea
  Malaysia                  malaysia             Malaysia
  Mexico                    mexico               Mexico
  New Zealand               newzealand           New Zealand
  Norway                    norway               Norway
  Poland                    poland               Poland
  Qatar                     qatar                Qatar
  Singapore                 singapore            Singapore
  South Africa              southafrica          South Africa
  Spain                     spain                Spain
  Sweden                    sweden               Sweden
  Switzerland               switzerland          Switzerland
  Taiwan                    taiwan               Taiwan
  United Arab Emirates      uae                  United Arab Emirates
  United Kingdom            uk                   United Kingdom
  United States             unitedstates         United States
  United States EUAP        unitedstateseuap     United States EUAP
  East Asia (Stage)         eastasiastage        (Asia Pacific) East Asia (Stage)
  Southeast Asia (Stage)    southeastasiastage   (Asia Pacific) Southeast Asia (Stage)
  Brazil US                 brazilus             (South America) Brazil US
  East US 2                 eastus2              (US) East US 2
  East US STG               eastusstg            (US) East US STG
  South Central US          southcentralus       (US) South Central US
  West US 3                 westus3              (US) West US 3
  North Central US          northcentralus       (US) North Central US
  West US                   westus               (US) West US
  Jio India West            jioindiawest         (Asia Pacific) Jio India West
  Central US EUAP           centraluseuap        (US) Central US EUAP
  South Central US STG      southcentralusstg    (US) South Central US STG
  West Central US           westcentralus        (US) West Central US
  South Africa West         southafricawest      (Africa) South Africa West
  Australia Central         australiacentral     (Asia Pacific) Australia Central
  Australia Central 2       australiacentral2    (Asia Pacific) Australia Central 2
  Australia Southeast       australiasoutheast   (Asia Pacific) Australia Southeast
  Jio India Central         jioindiacentral      (Asia Pacific) Jio India Central
  Korea South               koreasouth           (Asia Pacific) Korea South
  South India               southindia           (Asia Pacific) South India
  West India                westindia            (Asia Pacific) West India
  Canada East               canadaeast           (Canada) Canada East
  France South              francesouth          (Europe) France South
  Germany North             germanynorth         (Europe) Germany North
  Norway West               norwaywest           (Europe) Norway West
  Switzerland West          switzerlandwest      (Europe) Switzerland West
  UK West                   ukwest               (Europe) UK West
  UAE Central               uaecentral           (Middle East) UAE Central
  Brazil Southeast          brazilsoutheast      (South America) Brazil Southeast
  ```

</details>

### [`foundry-kinds.sh`](./queries/foundry-kinds.sh)

Outputs the available cognitive services account kinds. Note that the deployment models are either `OpenAI` or `AIServices`.

`--output` is the only supported argument.

Example:

```sh
. scripts/azure/queries/foundry-kinds.sh
```

<details>

  <summary>Output</summary>

  ```sh
  Result
  -----------------------------------
  AIServices
  AnomalyDetector
  CognitiveServices
  ComputerVision
  ContentModerator
  ContentSafety
  ConversationalLanguageUnderstanding
  CustomVision.Prediction
  CustomVision.Training
  Face
  FormRecognizer
  HealthInsights
  ImmersiveReader
  Internal.AllInOne
  LUIS.Authoring
  LanguageAuthoring
  MetricsAdvisor
  OpenAI
  Personalizer
  QnAMaker.v2
  SpeechServices
  TextAnalytics
  TextTranslation
  ```

</details>

### [`foundry-skus.sh`](./queries/foundry-skus.sh)

Outputs the available cognitive services account SKUs.

`--output` is the only supported argument.

Example:

```sh
. scripts/azure/queries/foundry-skus.sh
```

<details>

  <summary>Output</summary>

  ```sh
  Name    Tier        Kind
  ------  ----------  -----------------------------------
  C2      Standard    TextTranslation
  C3      Standard    TextTranslation
  C4      Standard    TextTranslation
  D3      Standard    TextTranslation
  E0      Enterprise  Face
  E0      Enterprise  Personalizer
  F0      Free        AnomalyDetector
  F0      Free        ComputerVision
  F0      Free        ContentModerator
  F0      Free        ContentSafety
  F0      Free        CustomVision.Prediction
  F0      Free        CustomVision.Training
  F0      Free        Face
  F0      Free        FormRecognizer
  F0      Free        HealthInsights
  F0      Free        LUIS.Authoring
  F0      Free        Personalizer
  F0      Free        SpeechServices
  F0      Free        TextAnalytics
  F0      Free        TextTranslation
  S       Standard    ConversationalLanguageUnderstanding
  S       Standard    LanguageAuthoring
  S       Standard    TextAnalytics
  S0      Standard    AIServices
  S0      Standard    AnomalyDetector
  S0      Standard    CognitiveServices
  S0      Standard    ContentModerator
  S0      Standard    ContentSafety
  S0      Standard    CustomVision.Prediction
  S0      Standard    CustomVision.Training
  S0      Standard    Face
  S0      Standard    FormRecognizer
  S0      Standard    HealthInsights
  S0      Standard    ImmersiveReader
  S0      Standard    Internal.AllInOne
  S0      Standard    MetricsAdvisor
  S0      Standard    OpenAI
  S0      Standard    Personalizer
  S0      Standard    QnAMaker.v2
  S0      Standard    SpeechServices
  S1      Standard    ComputerVision
  S1      Standard    ImmersiveReader
  S1      Standard    TextTranslation
  S2      Standard    TextTranslation
  S3      Standard    TextTranslation
  S4      Standard    TextTranslation
  ```

</details>

### [`ai-models.sh`](./queries/ai-models.sh)

Outputs the details necessary for planning a model deployment.

Argument | Description | Default
---------|-------------|--------
`--model` | Filter the available models by name (uses `contains`) | 
`--format` | Filter the available models by format (uses `contains`) | 
`--kind` | Filter the available models by kind (uses `contains`). Filters out `MaaS` if not provided. | 
`--location` | Region to find available models in | `eastus`
`--output` | The format to output results in | `table`

Example:

```sh
. scripts/azure/queries/ai-models.sh
```

<details>

  <summary>Output</summary>

  ```sh
  Kind        Name                                    Version           Format             SkuName             SkuCapacity
  ----------  --------------------------------------  ----------------  -----------------  ------------------  -------------
  OpenAI      dall-e-3                                3.0               OpenAI             Standard            1
  OpenAI      dall-e-2                                2.0               OpenAI             Standard            1
  OpenAI      code-cushman-fine-tune-002              1                 OpenAI             Standard            120
  OpenAI      gpt-35-turbo                            0301              OpenAI             Standard            120
  OpenAI      gpt-35-turbo                            0613              OpenAI             Standard            120
  OpenAI      gpt-35-turbo                            1106              OpenAI             GlobalBatch         10
  OpenAI      gpt-35-turbo                            0125              OpenAI             Standard            100
  OpenAI      gpt-35-turbo-instruct                   0914              OpenAI             Standard            120
  OpenAI      gpt-35-turbo-16k                        0613              OpenAI             Standard            120
  OpenAI      gpt-4                                   0125-Preview      OpenAI             Standard            10
  OpenAI      gpt-4                                   1106-Preview      OpenAI             ProvisionedManaged
  OpenAI      gpt-4                                   0613              OpenAI             Standard            10
  OpenAI      gpt-4-32k                               0613              OpenAI             ProvisionedManaged
  OpenAI      gpt-4                                   turbo-2024-04-09  OpenAI             Standard            10
  OpenAI      gpt-4o                                  2024-05-13        OpenAI             Standard            10
  OpenAI      gpt-4o                                  2024-08-06        OpenAI             Standard            10
  OpenAI      gpt-4o-mini                             2024-07-18        OpenAI             Standard            10
  OpenAI      gpt-4o                                  2024-11-20        OpenAI             Standard            10
  OpenAI      gpt-4o-audio-preview                    2024-12-17        OpenAI             Provisioned         100
  OpenAI      gpt-4o-mini-audio-preview               2024-12-17        OpenAI             GlobalStandard      100
  OpenAI      gpt-4.1                                 2025-04-14        OpenAI             GlobalStandard      10
  OpenAI      gpt-4.1-mini                            2025-04-14        OpenAI             GlobalStandard      10
  OpenAI      gpt-4.1-nano                            2025-04-14        OpenAI             DataZoneStandard    10
  OpenAI      o1-mini                                 2024-09-12        OpenAI             GlobalStandard      10
  OpenAI      o1                                      2024-12-17        OpenAI             Standard            10
  OpenAI      o3-mini                                 2025-01-31        OpenAI             GlobalStandard      10
  OpenAI      o4-mini                                 2025-04-16        OpenAI             GlobalStandard      10
  OpenAI      ada                                     1                 OpenAI
  OpenAI      text-ada-001                            1                 OpenAI             Standard            120
  OpenAI      text-similarity-ada-001                 1                 OpenAI             Standard            120
  OpenAI      text-embedding-ada-002                  1                 OpenAI             Standard            120
  OpenAI      text-embedding-ada-002                  2                 OpenAI             Standard            120
  OpenAI      babbage                                 1                 OpenAI
  OpenAI      text-babbage-001                        1                 OpenAI             Standard            120
  OpenAI      curie                                   1                 OpenAI
  OpenAI      text-curie-001                          1                 OpenAI             Standard            120
  OpenAI      text-similarity-curie-001               1                 OpenAI             Standard            120
  OpenAI      davinci                                 1                 OpenAI
  OpenAI      text-davinci-002                        1                 OpenAI             Standard            120
  OpenAI      text-davinci-003                        1                 OpenAI             Standard            60
  OpenAI      text-davinci-fine-tune-002              1                 OpenAI             Standard            120
  OpenAI      code-davinci-002                        1                 OpenAI             Standard            120
  OpenAI      code-davinci-fine-tune-002              1                 OpenAI             Standard            120
  OpenAI      text-embedding-3-small                  1                 OpenAI             Standard            120
  OpenAI      text-embedding-3-large                  1                 OpenAI             Standard            120
  AIServices  dall-e-3                                3.0               OpenAI             Standard            1
  AIServices  dall-e-2                                2.0               OpenAI             Standard            1
  AIServices  code-cushman-fine-tune-002              1                 OpenAI             Standard            120
  AIServices  gpt-35-turbo                            0301              OpenAI             Standard            120
  AIServices  gpt-35-turbo                            0613              OpenAI             Standard            120
  AIServices  gpt-35-turbo                            1106              OpenAI             GlobalBatch         10
  AIServices  gpt-35-turbo                            0125              OpenAI             Standard            100
  AIServices  gpt-35-turbo-instruct                   0914              OpenAI             Standard            120
  AIServices  gpt-35-turbo-16k                        0613              OpenAI             Standard            120
  AIServices  gpt-4                                   0125-Preview      OpenAI             Standard            10
  AIServices  gpt-4                                   1106-Preview      OpenAI             ProvisionedManaged
  AIServices  gpt-4                                   0613              OpenAI             Standard            10
  AIServices  gpt-4-32k                               0613              OpenAI             ProvisionedManaged
  AIServices  gpt-4                                   turbo-2024-04-09  OpenAI             Standard            10
  AIServices  gpt-4o                                  2024-05-13        OpenAI             Standard            10
  AIServices  gpt-4o                                  2024-08-06        OpenAI             Standard            10
  AIServices  gpt-4o-mini                             2024-07-18        OpenAI             Standard            10
  AIServices  gpt-4o                                  2024-11-20        OpenAI             Standard            10
  AIServices  gpt-4o-audio-preview                    2024-12-17        OpenAI             Provisioned         100
  AIServices  gpt-4o-mini-audio-preview               2024-12-17        OpenAI             GlobalStandard      100
  AIServices  gpt-4.1                                 2025-04-14        OpenAI             GlobalStandard      10
  AIServices  gpt-4.1-mini                            2025-04-14        OpenAI             GlobalStandard      10
  AIServices  gpt-4.1-nano                            2025-04-14        OpenAI             DataZoneStandard    10
  AIServices  o1-mini                                 2024-09-12        OpenAI             GlobalStandard      10
  AIServices  o1                                      2024-12-17        OpenAI             Standard            10
  AIServices  o3-mini                                 2025-01-31        OpenAI             GlobalStandard      10
  AIServices  o4-mini                                 2025-04-16        OpenAI             GlobalStandard      10
  AIServices  ada                                     1                 OpenAI
  AIServices  text-ada-001                            1                 OpenAI             Standard            120
  AIServices  text-similarity-ada-001                 1                 OpenAI             Standard            120
  AIServices  text-embedding-ada-002                  1                 OpenAI             Standard            120
  AIServices  text-embedding-ada-002                  2                 OpenAI             Standard            120
  AIServices  babbage                                 1                 OpenAI
  AIServices  text-babbage-001                        1                 OpenAI             Standard            120
  AIServices  curie                                   1                 OpenAI
  AIServices  text-curie-001                          1                 OpenAI             Standard            120
  AIServices  text-similarity-curie-001               1                 OpenAI             Standard            120
  AIServices  davinci                                 1                 OpenAI
  AIServices  text-davinci-002                        1                 OpenAI             Standard            120
  AIServices  text-davinci-003                        1                 OpenAI             Standard            60
  AIServices  text-davinci-fine-tune-002              1                 OpenAI             Standard            120
  AIServices  code-davinci-002                        1                 OpenAI             Standard            120
  AIServices  code-davinci-fine-tune-002              1                 OpenAI             Standard            120
  AIServices  text-embedding-3-small                  1                 OpenAI             Standard            120
  AIServices  text-embedding-3-large                  1                 OpenAI             Standard            120
  AIServices  AI21-Jamba-1.5-Large                    1                 AI21 Labs          GlobalStandard      1
  AIServices  AI21-Jamba-1.5-Mini                     1                 AI21 Labs          GlobalStandard      1
  AIServices  AI21-Jamba-Instruct                     1                 AI21 Labs          GlobalStandard      1
  AIServices  Codestral-2501                          2                 Mistral AI         GlobalStandard      1
  AIServices  cohere-command-a                        1                 Cohere             GlobalStandard      1
  AIServices  Cohere-command-r                        1                 Cohere             GlobalStandard      1
  AIServices  Cohere-command-r-08-2024                1                 Cohere             GlobalStandard      1
  AIServices  Cohere-command-r-plus                   1                 Cohere             GlobalStandard      1
  AIServices  Cohere-command-r-plus-08-2024           1                 Cohere             GlobalStandard      1
  AIServices  Cohere-embed-v3-english                 1                 Cohere             GlobalStandard      1
  AIServices  Cohere-embed-v3-multilingual            1                 Cohere             GlobalStandard      1
  AIServices  DeepSeek-R1                             1                 DeepSeek           GlobalStandard      5000
  AIServices  DeepSeek-R1-0528                        1                 DeepSeek           GlobalStandard      5000
  AIServices  DeepSeek-V3                             1                 DeepSeek           GlobalStandard      1
  AIServices  DeepSeek-V3-0324                        1                 DeepSeek           GlobalStandard      5000
  AIServices  embed-v-4-0                             1                 Cohere             GlobalStandard      1
  AIServices  jais-30b-chat                           1                 Core42             GlobalStandard      1
  AIServices  jais-30b-chat                           2                 Core42             GlobalStandard      1
  AIServices  jais-30b-chat                           3                 Core42             GlobalStandard      1
  AIServices  Llama-3.2-11B-Vision-Instruct           1                 Meta               GlobalStandard      1
  AIServices  Llama-3.2-11B-Vision-Instruct           2                 Meta               GlobalStandard      1
  AIServices  Llama-3.2-90B-Vision-Instruct           1                 Meta               GlobalStandard      1
  AIServices  Llama-3.2-90B-Vision-Instruct           2                 Meta               GlobalStandard      1
  AIServices  Llama-3.2-90B-Vision-Instruct           3                 Meta               GlobalStandard      1
  AIServices  Llama-3.3-70B-Instruct                  1                 Meta               GlobalStandard      1
  AIServices  Llama-3.3-70B-Instruct                  2                 Meta               GlobalStandard      1
  AIServices  Llama-3.3-70B-Instruct                  3                 Meta               GlobalStandard      1
  AIServices  Llama-3.3-70B-Instruct                  4                 Meta               GlobalStandard      1
  AIServices  Llama-3.3-70B-Instruct                  5                 Meta               GlobalStandard      1
  AIServices  MAI-DS-R1                               1                 Microsoft          GlobalStandard      5000
  AIServices  Meta-Llama-3-70B-Instruct               6                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3-70B-Instruct               7                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3-70B-Instruct               8                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3-70B-Instruct               9                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3-8B-Instruct                6                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3-8B-Instruct                7                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3-8B-Instruct                8                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3-8B-Instruct                9                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-405B-Instruct            1                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-70B-Instruct             1                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-70B-Instruct             2                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-70B-Instruct             3                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-70B-Instruct             4                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-8B-Instruct              1                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-8B-Instruct              2                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-8B-Instruct              3                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-8B-Instruct              4                 Meta               GlobalStandard      1
  AIServices  Meta-Llama-3.1-8B-Instruct              5                 Meta               GlobalStandard      1
  AIServices  Ministral-3B                            1                 Mistral AI         GlobalStandard      1
  AIServices  Mistral-large-2407                      1                 Mistral AI         GlobalStandard      1
  AIServices  Mistral-Large-2411                      2                 Mistral AI         GlobalStandard      1
  AIServices  Mistral-Nemo                            1                 Mistral AI         GlobalStandard      1
  AIServices  Mistral-small                           1                 Mistral AI         GlobalStandard      1
  AIServices  mistral-small-2503                      1                 Mistral AI         GlobalStandard      1
  AIServices  Phi-3-medium-128k-instruct              3                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-medium-128k-instruct              4                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-medium-128k-instruct              5                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-medium-128k-instruct              6                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-medium-128k-instruct              7                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-medium-4k-instruct                3                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-medium-4k-instruct                4                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-medium-4k-instruct                5                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-medium-4k-instruct                6                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-128k-instruct                10                Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-128k-instruct                11                Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-128k-instruct                12                Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-128k-instruct                13                Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-4k-instruct                  10                Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-4k-instruct                  11                Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-4k-instruct                  13                Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-4k-instruct                  14                Microsoft          GlobalStandard      1
  AIServices  Phi-3-mini-4k-instruct                  15                Microsoft          GlobalStandard      1
  AIServices  Phi-3-small-128k-instruct               3                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-small-128k-instruct               4                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-small-128k-instruct               5                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-small-8k-instruct                 3                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-small-8k-instruct                 4                 Microsoft          GlobalStandard      1
  AIServices  Phi-3-small-8k-instruct                 5                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-mini-instruct                   1                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-mini-instruct                   2                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-mini-instruct                   3                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-mini-instruct                   4                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-mini-instruct                   6                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-MoE-instruct                    2                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-MoE-instruct                    3                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-MoE-instruct                    4                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-MoE-instruct                    5                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-vision-instruct                 1                 Microsoft          GlobalStandard      1
  AIServices  Phi-3.5-vision-instruct                 2                 Microsoft          GlobalStandard      1
  AIServices  Phi-4                                   2                 Microsoft          GlobalStandard      1
  AIServices  Phi-4                                   3                 Microsoft          GlobalStandard      1
  AIServices  Phi-4                                   4                 Microsoft          GlobalStandard      1
  AIServices  Phi-4                                   5                 Microsoft          GlobalStandard      1
  AIServices  Phi-4                                   6                 Microsoft          GlobalStandard      1
  AIServices  Phi-4                                   7                 Microsoft          GlobalStandard      1
  AIServices  Phi-4-mini-instruct                     1                 Microsoft          GlobalStandard      1
  AIServices  Phi-4-multimodal-instruct               1                 Microsoft          GlobalStandard      1
  AIServices  Phi-4-reasoning                         1                 Microsoft          GlobalStandard      1
  AIServices  Phi-4-mini-reasoning                    1                 Microsoft          GlobalStandard      1
  AIServices  Llama-4-Maverick-17B-128E-Instruct-FP8  1                 Meta               GlobalStandard      1
  AIServices  Llama-4-Scout-17B-16E-Instruct          1                 Meta               GlobalStandard      1
  AIServices  mistral-medium-2505                     1                 Mistral AI         GlobalStandard      1
  AIServices  mistral-document-ai-2505                1                 Mistral AI         GlobalStandard      10
  AIServices  grok-3                                  1                 xAI                GlobalStandard      10
  AIServices  grok-3-mini                             1                 xAI                GlobalStandard      10
  AIServices  FLUX-1.1-pro                            1                 Black Forest Labs  GlobalStandard      1
  AIServices  FLUX.1-Kontext-pro                      1                 Black Forest Labs  GlobalStandard      1
  AIServices  gpt-oss-120b                            1                 OpenAI-OSS         GlobalStandard      500
  ```

</details>

## References

- [Create and Deploy an Azure OpenAI in Azure AI Foundry Models Resource](https://learn.microsoft.com/en-us/azure/ai-foundry/openai/how-to/create-resource?pivots=cli)
- [Develop with AI Models](https://learn.microsoft.com/en-us/azure/ai-foundry/foundry-models/concepts/endpoints?tabs=rest)
- [Azure OpenAI API Version Support](https://learn.microsoft.com/en-us/azure/ai-foundry/openai/supported-languages?source=recommendations&tabs=dotnet-secure%2Csecure%2Cpython-key%2Ccommand&pivots=programming-language-go)
- [`az cognitiveservices account`](https://learn.microsoft.com/en-us/cli/azure/cognitiveservices/account?view=azure-cli-latest&preserve-view=true#az-cognitiveservices-account-create)
- [`az cognitiveservices account deployment`](https://learn.microsoft.com/en-us/cli/azure/cognitiveservices/account/deployment?view=azure-cli-latest&preserve-view=true#az-cognitiveservices-account-deployment-create)
