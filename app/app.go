package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	tmdb "github.com/cometbft/cometbft-db"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/x/consensus"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"

	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"

	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"

	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/cosmos/cosmos-sdk/x/nft"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	nftmodule "github.com/cosmos/cosmos-sdk/x/nft/module"

	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramsproposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	alliancebank "github.com/terra-money/alliance/custom/bank"
	alliancebankkeeper "github.com/terra-money/alliance/custom/bank/keeper"
	alliance "github.com/terra-money/alliance/x/alliance"
	allianceclient "github.com/terra-money/alliance/x/alliance/client"
	alliancekeeper "github.com/terra-money/alliance/x/alliance/keeper"
	alliancetypes "github.com/terra-money/alliance/x/alliance/types"

	"github.com/terra-money/feather-core/app/openapiconsole"
	appparams "github.com/terra-money/feather-core/app/params"
	"github.com/terra-money/feather-core/docs"
)

// DO NOT change the names of these variables!
// TODO: to prevent other users from changing these variables, we could probably just publish our own package like https://pkg.go.dev/github.com/cosmos/cosmos-sdk/version
var (
	AccountAddressPrefix       = "feath"
	AccountPubKeyPrefix        = "feathpub"
	ValidatorAddressPrefix     = "feathvaloper"
	ValidatorPubKeyPrefix      = "feathvaloperpub"
	ConsensusNodeAddressPrefix = "feathvalcons"
	ConsensusNodePubKeyPrefix  = "feathvalconspub"
	BondDenom                  = "featherstake"
	AppName                    = "feather-core"
)

// TODO: What is this?
func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	// this line is used by starport scaffolding # stargate/app/govProposalHandlers

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		allianceclient.CreateAllianceProposalHandler,
		allianceclient.UpdateAllianceProposalHandler,
		allianceclient.DeleteAllianceProposalHandler,
		// this line is used by starport scaffolding # stargate/app/govProposalHandler
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		capability.AppModuleBasic{},
		consensus.AppModuleBasic{},
		crisis.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		nftmodule.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		vesting.AppModuleBasic{},
		slashing.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()), // TODO: Do we need the legacy proposal handlers?
		distr.AppModuleBasic{},
		params.AppModuleBasic{},
		evidence.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		wasm.AppModuleBasic{},
		alliance.AppModuleBasic{},
	)
)

var (
	_ servertypes.Application = (*App)(nil)
	_ runtime.AppI            = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+AppName)
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	cdc               codec.Codec
	legacyAmino       *codec.LegacyAmino
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	keys map[string]*storetypes.KVStoreKey

	// keepers
	AuthKeeper            authkeeper.AccountKeeper // TODO: Do we even need to store this state?
	AuthzKeeper           authzkeeper.Keeper
	BankKeeper            alliancebankkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	NftKeeper             nftkeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper
	ICAHostKeeper         icahostkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	WasmKeeper            wasm.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	AllianceKeeper        alliancekeeper.Keeper

	// ModuleManager is the module manager
	ModuleManager *module.Manager

	// sm is the simulation manager
	sm           *module.SimulationManager
	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db tmdb.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	cdc := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	// Init App
	app := &App{
		BaseApp: baseapp.NewBaseApp(
			AppName,
			logger,
			db,
			encodingConfig.TxConfig.TxDecoder(),
			baseAppOptions...,
		),
		cdc:               cdc,
		legacyAmino:       legacyAmino,
		interfaceRegistry: interfaceRegistry,
		txConfig:          txConfig,
		keys:              make(map[string]*storetypes.KVStoreKey),
	}
	defer app.Seal()
	app.SetCommitMultiStoreTracer(traceStore)
	app.SetVersion(version.Version)
	app.SetInterfaceRegistry(interfaceRegistry)

	var modules []module.AppModule = make([]module.AppModule, 0)
	var simModules []module.AppModuleSimulation = make([]module.AppModuleSimulation, 0)

	// 'auth' module
	app.keys[authtypes.StoreKey] = storetypes.NewKVStoreKey(authtypes.StoreKey)
	app.MountStores(app.keys[authtypes.StoreKey])
	app.AuthKeeper = authkeeper.NewAccountKeeper(
		cdc,
		app.keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		make(map[string][]string), // This will be populated by each module later
		AccountAddressPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	defer func() { // TODO: Does deferring this even work?
		app.AuthKeeper.GetModulePermissions()[authtypes.FeeCollectorName] = authtypes.NewPermissionsForAddress(authtypes.FeeCollectorName, nil) // This implicitly creates a module account
		app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(authtypes.FeeCollectorName).String()] = true
	}()
	modules = append(modules, auth.NewAppModule(cdc, app.AuthKeeper, nil, nil))
	simModules = append(simModules, auth.NewAppModule(cdc, app.AuthKeeper, authsim.RandomGenesisAccounts, nil))

	// 'bank' module - depends on
	// 1. 'auth'
	app.keys[banktypes.StoreKey] = storetypes.NewKVStoreKey(banktypes.StoreKey)
	app.MountStores(app.keys[banktypes.StoreKey])
	app.BankKeeper = alliancebankkeeper.NewBaseKeeper( // Use 'alliance' module's custom implementation instead
		cdc,
		app.keys[banktypes.StoreKey],
		app.AuthKeeper,
		make(map[string]bool),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	modules = append(modules, alliancebank.NewAppModule(cdc, app.BankKeeper, app.AuthKeeper, nil))
	simModules = append(simModules, alliancebank.NewAppModule(cdc, app.BankKeeper, app.AuthKeeper, nil))

	// 'authz' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.keys[authzkeeper.StoreKey] = storetypes.NewKVStoreKey(authzkeeper.StoreKey)
	app.MountStores(app.keys[authzkeeper.StoreKey])
	app.AuthzKeeper = authzkeeper.NewKeeper(
		app.keys[authzkeeper.StoreKey],
		cdc,
		app.MsgServiceRouter(), // TODO: Find out what this is
		app.AuthKeeper,
	)
	modules = append(modules, authzmodule.NewAppModule(cdc, app.AuthzKeeper, app.AuthKeeper, app.BankKeeper, interfaceRegistry))
	simModules = append(simModules, authzmodule.NewAppModule(cdc, app.AuthzKeeper, app.AuthKeeper, app.BankKeeper, interfaceRegistry))

	// 'capability' module
	app.keys[capabilitytypes.StoreKey] = storetypes.NewKVStoreKey(capabilitytypes.StoreKey)
	capabilityMemStoreKey := storetypes.NewMemoryStoreKey(capabilitytypes.MemStoreKey)
	app.MountStores(app.keys[capabilitytypes.StoreKey], capabilityMemStoreKey)
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		cdc,
		app.keys[capabilitytypes.StoreKey],
		capabilityMemStoreKey,
	)
	defer app.CapabilityKeeper.Seal()
	modules = append(modules, capability.NewAppModule(cdc, *app.CapabilityKeeper, false)) // TODO: Find out what is sealkeeper
	simModules = append(simModules, capability.NewAppModule(cdc, *app.CapabilityKeeper, false))

	// 'consensus' module
	app.keys[consensustypes.StoreKey] = storetypes.NewKVStoreKey(consensustypes.StoreKey)
	app.MountStores(app.keys[consensustypes.StoreKey])
	app.ConsensusParamsKeeper = consensuskeeper.NewKeeper(
		cdc,
		app.keys[consensustypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SetParamStore(&app.ConsensusParamsKeeper)
	modules = append(modules, consensus.NewAppModule(cdc, app.ConsensusParamsKeeper))

	// 'crisis' module - depends on
	// 1. 'bank'
	app.keys[crisistypes.StoreKey] = storetypes.NewKVStoreKey(crisistypes.StoreKey)
	app.MountStores(app.keys[crisistypes.StoreKey])
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		cdc,
		app.keys[crisistypes.StoreKey],
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	modules = append(modules, crisis.NewAppModule(app.CrisisKeeper, false, nil)) // Never skip invariant checks on genesis
	defer func() { app.ModuleManager.RegisterInvariants(app.CrisisKeeper) }()

	// 'feegrant' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.keys[feegrant.StoreKey] = storetypes.NewKVStoreKey(feegrant.StoreKey)
	app.MountStores(app.keys[feegrant.StoreKey])
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		cdc,
		app.keys[feegrant.StoreKey],
		app.AuthKeeper,
	)
	modules = append(modules, feegrantmodule.NewAppModule(cdc, app.AuthKeeper, app.BankKeeper, app.FeeGrantKeeper, interfaceRegistry))
	simModules = append(simModules, feegrantmodule.NewAppModule(cdc, app.AuthKeeper, app.BankKeeper, app.FeeGrantKeeper, interfaceRegistry))

	// 'group' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.keys[group.StoreKey] = storetypes.NewKVStoreKey(group.StoreKey)
	app.MountStores(app.keys[group.StoreKey])
	app.GroupKeeper = groupkeeper.NewKeeper(
		app.keys[group.StoreKey],
		cdc,
		app.MsgServiceRouter(),
		app.AuthKeeper,
		group.DefaultConfig(),
	)
	modules = append(modules, groupmodule.NewAppModule(cdc, app.GroupKeeper, app.AuthKeeper, app.BankKeeper, interfaceRegistry))
	simModules = append(simModules, groupmodule.NewAppModule(cdc, app.GroupKeeper, app.AuthKeeper, app.BankKeeper, interfaceRegistry))

	// 'staking' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.AuthKeeper.GetModulePermissions()[stakingtypes.BondedPoolName] = authtypes.NewPermissionsForAddress(stakingtypes.BondedPoolName, []string{authtypes.Burner, authtypes.Staking})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String()] = true
	app.AuthKeeper.GetModulePermissions()[stakingtypes.NotBondedPoolName] = authtypes.NewPermissionsForAddress(stakingtypes.NotBondedPoolName, []string{authtypes.Burner, authtypes.Staking})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String()] = true
	app.keys[stakingtypes.StoreKey] = storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	app.MountStores(app.keys[stakingtypes.StoreKey])
	app.StakingKeeper = stakingkeeper.NewKeeper(
		cdc,
		app.keys[stakingtypes.StoreKey],
		app.AuthKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	var stakingHooks []stakingtypes.StakingHooks = make([]stakingtypes.StakingHooks, 0)
	defer func() { app.StakingKeeper.SetHooks(stakingtypes.NewMultiStakingHooks(stakingHooks...)) }()
	modules = append(modules, staking.NewAppModule(cdc, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, nil))
	simModules = append(simModules, staking.NewAppModule(cdc, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, nil))

	// 'mint' module - depends on
	// 1. 'staking'
	// 2. 'auth'
	// 3. 'bank'
	app.AuthKeeper.GetModulePermissions()[minttypes.ModuleName] = authtypes.NewPermissionsForAddress(minttypes.ModuleName, []string{authtypes.Minter})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(minttypes.ModuleName).String()] = true
	app.keys[minttypes.StoreKey] = storetypes.NewKVStoreKey(minttypes.StoreKey)
	app.MountStores(app.keys[minttypes.StoreKey])
	app.MintKeeper = mintkeeper.NewKeeper(
		cdc,
		app.keys[minttypes.StoreKey],
		app.StakingKeeper,
		app.AuthKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName, // TODO: Find out what this is
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	modules = append(modules, mint.NewAppModule(cdc, app.MintKeeper, app.AuthKeeper, nil, nil))
	simModules = append(simModules, mint.NewAppModule(cdc, app.MintKeeper, app.AuthKeeper, nil, nil))

	// 'nft' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	app.AuthKeeper.GetModulePermissions()[nft.ModuleName] = authtypes.NewPermissionsForAddress(nft.ModuleName, nil)
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(nft.ModuleName).String()] = true
	app.keys[nftkeeper.StoreKey] = storetypes.NewKVStoreKey(nftkeeper.StoreKey)
	app.MountStores(app.keys[nftkeeper.StoreKey])
	app.NftKeeper = nftkeeper.NewKeeper(
		app.keys[nftkeeper.StoreKey],
		cdc,
		app.AuthKeeper,
		app.BankKeeper,
	)
	modules = append(modules, nftmodule.NewAppModule(cdc, app.NftKeeper, app.AuthKeeper, app.BankKeeper, interfaceRegistry))
	simModules = append(simModules, nftmodule.NewAppModule(cdc, app.NftKeeper, app.AuthKeeper, app.BankKeeper, interfaceRegistry))

	// 'genutil' module - depends on
	// 1. 'auth'
	// 2. 'staking'
	modules = append(modules, genutil.NewAppModule(app.AuthKeeper, app.StakingKeeper, app.BaseApp.DeliverTx, encodingConfig.TxConfig))

	// 'vesting' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	modules = append(modules, vesting.NewAppModule(app.AuthKeeper, app.BankKeeper))

	// 'slashing' module - depends on
	// 1. 'staking'
	// 2. 'auth'
	// 3. 'bank'
	app.keys[slashingtypes.StoreKey] = storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	app.MountStores(app.keys[slashingtypes.StoreKey])
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		cdc,
		encodingConfig.Amino,
		app.keys[slashingtypes.StoreKey],
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	stakingHooks = append(stakingHooks, app.SlashingKeeper.Hooks())
	modules = append(modules, slashing.NewAppModule(cdc, app.SlashingKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, nil))
	simModules = append(simModules, slashing.NewAppModule(cdc, app.SlashingKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, nil))

	// 'gov' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	// 3. 'staking'
	app.AuthKeeper.GetModulePermissions()[govtypes.ModuleName] = authtypes.NewPermissionsForAddress(govtypes.ModuleName, []string{authtypes.Burner})
	app.keys[govtypes.StoreKey] = storetypes.NewKVStoreKey(govtypes.StoreKey)
	app.MountStores(app.keys[govtypes.StoreKey])
	app.GovKeeper = *govkeeper.NewKeeper(
		cdc,
		app.keys[govtypes.StoreKey],
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	// Set legacy router for backwards compatibility with gov v1beta1
	govLegacyRouter := govv1beta1.NewRouter()
	defer app.GovKeeper.SetLegacyRouter(govLegacyRouter)
	govLegacyRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler)
	modules = append(modules, gov.NewAppModule(cdc, &app.GovKeeper, app.AuthKeeper, app.BankKeeper, nil))
	simModules = append(simModules, gov.NewAppModule(cdc, &app.GovKeeper, app.AuthKeeper, app.BankKeeper, nil))

	// 'distribution' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	// 3. 'staking'
	// 4. 'gov'
	app.AuthKeeper.GetModulePermissions()[distrtypes.ModuleName] = authtypes.NewPermissionsForAddress(distrtypes.ModuleName, nil)
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(distrtypes.ModuleName).String()] = true
	app.keys[distrtypes.StoreKey] = storetypes.NewKVStoreKey(distrtypes.StoreKey)
	app.MountStores(app.keys[distrtypes.StoreKey])
	app.DistrKeeper = distrkeeper.NewKeeper(
		cdc,
		app.keys[distrtypes.StoreKey],
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	stakingHooks = append(stakingHooks, app.DistrKeeper.Hooks())
	modules = append(modules, distr.NewAppModule(cdc, app.DistrKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, nil))
	simModules = append(simModules, distr.NewAppModule(cdc, app.DistrKeeper, app.AuthKeeper, app.BankKeeper, app.StakingKeeper, nil))

	// 'params' module - depends on
	// 1. 'gov'
	app.keys[paramstypes.StoreKey] = storetypes.NewKVStoreKey(paramstypes.StoreKey)
	paramsTStoreKey := storetypes.NewTransientStoreKey(paramstypes.TStoreKey)
	app.MountStores(app.keys[paramstypes.StoreKey], paramsTStoreKey)
	app.ParamsKeeper = paramskeeper.NewKeeper(
		cdc,
		legacyAmino,
		app.keys[paramstypes.StoreKey],
		paramsTStoreKey,
	)
	govLegacyRouter.AddRoute(paramsproposaltypes.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper))
	modules = append(modules, params.NewAppModule(app.ParamsKeeper))
	simModules = append(simModules, params.NewAppModule(app.ParamsKeeper))

	// 'evidence' module - depends on
	// 1. 'staking'
	// 2. 'slashing'
	app.keys[evidencetypes.StoreKey] = storetypes.NewKVStoreKey(evidencetypes.StoreKey)
	app.MountStores(app.keys[evidencetypes.StoreKey])
	app.EvidenceKeeper = *evidencekeeper.NewKeeper(
		cdc,
		app.keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)
	modules = append(modules, evidence.NewAppModule(app.EvidenceKeeper))
	simModules = append(simModules, evidence.NewAppModule(app.EvidenceKeeper))

	// 'upgrade' module - depends on
	// 1. 'gov'
	app.keys[upgradetypes.StoreKey] = storetypes.NewKVStoreKey(upgradetypes.StoreKey)
	app.MountStores(app.keys[upgradetypes.StoreKey])
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights, // TODO: What is this?
		app.keys[upgradetypes.StoreKey],
		cdc,
		homePath,
		app,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	govLegacyRouter.AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper))
	modules = append(modules, upgrade.NewAppModule(app.UpgradeKeeper))

	// 'ibc' module - depends on
	// 1. 'staking'
	// 2. 'upgrade'
	// 3. 'capability'
	// 4. 'gov'
	// 5. 'params'
	app.keys[ibcexported.StoreKey] = storetypes.NewKVStoreKey(ibcexported.StoreKey)
	app.MountStores(app.keys[ibcexported.StoreKey])
	app.IBCKeeper = ibckeeper.NewKeeper(
		cdc,
		app.keys[ibcexported.StoreKey],
		app.ParamsKeeper.Subspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName),
	)
	// app.IBCKeeper.SetRouter(ibcporttypes.NewRouter())
	govLegacyRouter.AddRoute(ibcexported.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper))
	modules = append(modules, ibc.NewAppModule(app.IBCKeeper))
	simModules = append(simModules, ibc.NewAppModule(app.IBCKeeper))

	// 'ibctransfer' module - depends on
	// 1. 'ibc'
	// 2. 'auth'
	// 3. 'bank'
	// 4. 'capability'
	app.AuthKeeper.GetModulePermissions()[ibctransfertypes.ModuleName] = authtypes.NewPermissionsForAddress(ibctransfertypes.ModuleName, []string{authtypes.Minter, authtypes.Burner})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(ibctransfertypes.ModuleName).String()] = true
	app.keys[ibctransfertypes.StoreKey] = storetypes.NewKVStoreKey(ibctransfertypes.StoreKey)
	app.MountStores(app.keys[ibctransfertypes.StoreKey])
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		cdc,
		app.keys[ibctransfertypes.StoreKey],
		app.ParamsKeeper.Subspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AuthKeeper,
		app.BankKeeper,
		app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName),
	)
	// app.IBCKeeper.Router.AddRoute(ibctransfertypes.ModuleName, transfer.NewIBCModule(app.TransferKeeper))
	modules = append(modules, ibctransfer.NewAppModule(app.TransferKeeper))
	simModules = append(simModules, ibctransfer.NewAppModule(app.TransferKeeper))

	// 'ica'
	app.AuthKeeper.GetModulePermissions()[icatypes.ModuleName] = authtypes.NewPermissionsForAddress(icatypes.ModuleName, nil)
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(icatypes.ModuleName).String()] = true

	// 'icacontroller' module - depends on
	// 1. 'ibc'
	// 2. 'capability'
	app.keys[icacontrollertypes.StoreKey] = storetypes.NewKVStoreKey(icacontrollertypes.StoreKey)
	app.MountStores(app.keys[icacontrollertypes.StoreKey])
	icaControllerKeeper := icacontrollerkeeper.NewKeeper(
		cdc,
		app.keys[icacontrollertypes.StoreKey],
		app.ParamsKeeper.Subspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 fee
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName),
		app.MsgServiceRouter(),
	)

	// 'icahost' module - depends on
	// 1. 'ibc'
	// 2. 'auth'
	// 3. 'capability'
	// 4. 'icacontroller'
	app.keys[icahosttypes.StoreKey] = storetypes.NewKVStoreKey(icahosttypes.StoreKey)
	app.MountStores(app.keys[icahosttypes.StoreKey])
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		cdc,
		app.keys[icahosttypes.StoreKey],
		app.ParamsKeeper.Subspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AuthKeeper,
		app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName),
		app.MsgServiceRouter(),
	)
	// app.IBCKeeper.Router.AddRoute(icahosttypes.SubModuleName, icahost.NewIBCModule(app.ICAHostKeeper))
	modules = append(modules, ica.NewAppModule(&icaControllerKeeper, &app.ICAHostKeeper))
	simModules = append(simModules, ica.NewAppModule(&icaControllerKeeper, &app.ICAHostKeeper))

	// 'wasm' module - depends on
	// 1. 'gov'
	// 2. 'auth'
	// 3. 'bank'
	// 4. 'staking'
	// 5. 'distribution'
	// 6. 'capability'
	// 7. 'ibc'
	// 8. 'ibctransfer'
	app.AuthKeeper.GetModulePermissions()[wasmtypes.ModuleName] = authtypes.NewPermissionsForAddress(wasmtypes.ModuleName, []string{authtypes.Burner})
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(wasmtypes.ModuleName).String()] = true
	app.keys[wasmtypes.StoreKey] = storetypes.NewKVStoreKey(wasmtypes.StoreKey)
	app.MountStores(app.keys[wasmtypes.StoreKey])
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}
	app.WasmKeeper = wasm.NewKeeper(
		cdc,
		app.keys[wasmtypes.StoreKey],
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.CapabilityKeeper.ScopeToModule(wasm.ModuleName),
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		filepath.Join(homePath, "wasm"),
		wasmConfig,
		"iterator,staking,stargate,cosmwasm_1_1", // TODO: Find out what this configures
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	govLegacyRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(app.WasmKeeper, wasm.EnableAllProposals))
	modules = append(modules, wasm.NewAppModule(cdc, &app.WasmKeeper, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, app.MsgServiceRouter(), nil))
	simModules = append(simModules, wasm.NewAppModule(cdc, &app.WasmKeeper, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, app.MsgServiceRouter(), nil))

	// 'alliance' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	// 3. 'staking'
	// 4. 'distribution'
	// 5. 'gov'
	app.BankKeeper.RegisterKeepers(app.AllianceKeeper, app.StakingKeeper)
	app.AuthKeeper.GetModulePermissions()[alliancetypes.ModuleName] = authtypes.NewPermissionsForAddress(alliancetypes.ModuleName, []string{authtypes.Minter, authtypes.Burner})
	app.AuthKeeper.GetModulePermissions()[alliancetypes.RewardsPoolName] = authtypes.NewPermissionsForAddress(alliancetypes.RewardsPoolName, nil)
	app.BankKeeper.GetBlockedAddresses()[authtypes.NewModuleAddress(alliancetypes.RewardsPoolName).String()] = true
	app.keys[alliancetypes.StoreKey] = storetypes.NewKVStoreKey(alliancetypes.StoreKey)
	app.MountStores(app.keys[alliancetypes.StoreKey])
	app.AllianceKeeper = alliancekeeper.NewKeeper(
		cdc,
		app.keys[alliancetypes.StoreKey],
		app.ParamsKeeper.Subspace(alliancetypes.ModuleName),
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
	)
	govLegacyRouter.AddRoute(alliancetypes.RouterKey, alliance.NewAllianceProposalHandler(app.AllianceKeeper))
	stakingHooks = append(stakingHooks, app.AllianceKeeper.StakingHooks())
	modules = append(modules, alliance.NewAppModule(cdc, app.AllianceKeeper, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, interfaceRegistry))
	simModules = append(simModules, alliance.NewAppModule(cdc, app.AllianceKeeper, app.StakingKeeper, app.AuthKeeper, app.BankKeeper, interfaceRegistry))

	/****  Module Options ****/

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.ModuleManager = module.NewManager(modules...)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
		// upgrades should be run first
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		consensustypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		wasm.ModuleName,
		alliancetypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
	)

	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		consensustypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		wasm.ModuleName,
		alliancetypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/endBlockers
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.ModuleManager.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		consensustypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		wasm.ModuleName,
		alliancetypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/initGenesis
	)

	// Uncomment if you want to set a custom migration order here.
	// app.mm.SetOrderMigrations(custom order)

	app.configurator = module.NewConfigurator(cdc, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.ModuleManager.RegisterServices(app.configurator)

	// create the simulation manager and define the order of the modules for deterministic simulations
	app.sm = module.NewSimulationManager(simModules...)
	app.sm.RegisterStoreDecoders()

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AuthKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				FeegrantKeeper:  app.FeeGrantKeeper,
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			IBCKeeper:         app.IBCKeeper,
			TxCounterStoreKey: app.keys[wasmtypes.StoreKey],
			WasmConfig:        wasmConfig,
			Cdc:               cdc,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			logger.Error("error on loading last version", "err", err)
			os.Exit(1)
		}
	}

	return app
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdktypes.Context, req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	return app.ModuleManager.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdktypes.Context, req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return app.ModuleManager.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdktypes.Context, req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	return app.ModuleManager.InitGenesis(ctx, app.cdc, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.cdc
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns a TxConfig
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/openapi.yml", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapiconsole.Handler(AppName, "/static/openapi.yml"))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *App) RegisterNodeService(clientCtx client.Context) {
	node.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return ModuleBasics.DefaultGenesis(app.cdc)
}
