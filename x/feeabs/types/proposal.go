package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	_ govtypes.Content = &AddHostZoneProposal{}
)

const (
	// ProposalTypeAddHostZone defines the type for a AddHostZoneProposal
	ProposalTypeAddHostZone = "AddHostZone"
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddHostZone)
}

// NewClientUpdateProposal creates a new client update proposal.
func NewAddHostZoneProposal(title, description string, config HostChainFeeAbsConfig) govtypes.Content {
	return &AddHostZoneProposal{
		Title:           title,
		Description:     description,
		HostChainConfig: &config,
	}
}

// GetTitle returns the title of a client update proposal.
func (ahzp *AddHostZoneProposal) GetTitle() string { return ahzp.Title }

// GetDescription returns the description of a client update proposal.
func (ahzp *AddHostZoneProposal) GetDescription() string { return ahzp.Description }

// ProposalRoute returns the routing key of a client update proposal.
func (ahzp *AddHostZoneProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a client update proposal.
func (ahzp *AddHostZoneProposal) ProposalType() string { return ProposalTypeAddHostZone }

// ValidateBasic runs basic stateless validity checks
func (ahzp *AddHostZoneProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(ahzp)
	if err != nil {
		return err
	}

	// TODO: add validate here

	return nil
}
